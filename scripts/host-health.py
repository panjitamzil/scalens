#!/usr/bin/env python3
import argparse
import json
import os
import platform
import re
import subprocess
import time
from pathlib import Path


def command(*args):
    return subprocess.check_output(args, text=True, timeout=10)


def linux_snapshot(proc=Path("/proc")):
    memory = {}
    for line in (proc / "meminfo").read_text().splitlines():
        key, value, *_ = line.split()
        memory[key.rstrip(":")] = int(value) * 1024
    vm = dict(line.split() for line in (proc / "vmstat").read_text().splitlines())
    cpu = [int(v) for v in (proc / "stat").read_text().splitlines()[0].split()[1:9]]
    psi = {}
    for name in ("cpu", "memory", "io"):
        path = proc / "pressure" / name
        if path.exists():
            psi[name] = {
                line.split()[0]: {k: float(v) for k, v in (item.split("=") for item in line.split()[1:])}
                for line in path.read_text().splitlines()
            }
    return {
        "memory_total_bytes": memory["MemTotal"],
        "memory_available_bytes": memory["MemAvailable"],
        "swap_used_bytes": memory.get("SwapTotal", 0) - memory.get("SwapFree", 0),
        "swapout_pages": int(vm["pswpout"]),
        "cpu_ticks": cpu,
        "pressure": psi,
        "boot_id": (proc / "sys/kernel/random/boot_id").read_text().strip(),
    }


def snapshot():
    result = {"timestamp": time.time(), "os": platform.system(), "architecture": platform.machine(),
              "cpu_count": os.cpu_count(), "page_size": os.sysconf("SC_PAGE_SIZE")}
    try:
        if result["os"] == "Linux":
            result.update(linux_snapshot())
        elif result["os"] == "Darwin":
            vm = command("vm_stat")
            result.update(swapout_pages=int(re.search(r"Swapouts:\s+(\d+)", vm)[1]),
                          vm_stat=vm, swap=command("sysctl", "vm.swapusage"),
                          cpu=command("top", "-l", "1", "-n", "0"),
                          boot_id=command("sysctl", "-n", "kern.boottime").strip())
        else:
            result["error"] = "Host health is not supported on this OS"
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        result["error"] = str(error)
    return result


def assess(samples):
    thresholds = {"swapout_pages": 1024, "cpu_busy_fraction": .9, "cpu_steal_fraction": .05,
                  "memory_available_fraction": .05, "memory_full_psi_fraction": .01}
    reasons, unknown = [], []
    if len(samples) < 2 or any(s.get("error") for s in samples):
        return {"status": "unverified_environment", "reasons": ["host samples missing or failed"],
                "thresholds": thresholds}
    for before, after in zip(samples, samples[1:]):
        elapsed = after["timestamp"] - before["timestamp"]
        if elapsed <= 0 or before.get("boot_id") != after.get("boot_id"):
            unknown.append("host restarted or timestamps are invalid")
            continue
        if after["swapout_pages"] < before["swapout_pages"]:
            unknown.append("swap counter reset")
        if after["os"] != "Linux":
            continue
        ticks = [b - a for a, b in zip(before["cpu_ticks"], after["cpu_ticks"])]
        total = sum(ticks)
        if total <= 0 or min(ticks) < 0:
            unknown.append("CPU counters missing or reset")
        else:
            if 1 - (ticks[3] + ticks[4]) / total > thresholds["cpu_busy_fraction"]:
                reasons.append("Linux CPU busy exceeded 90% between samples")
            if ticks[7] / total > thresholds["cpu_steal_fraction"]:
                reasons.append("Linux CPU steal exceeded 5% between samples")
        if after["memory_available_bytes"] / after["memory_total_bytes"] < thresholds["memory_available_fraction"]:
            reasons.append("Linux available memory below 5%")
        try:
            stalled = (after["pressure"]["memory"]["full"]["total"] -
                       before["pressure"]["memory"]["full"]["total"]) / (elapsed * 1_000_000)
            if stalled > thresholds["memory_full_psi_fraction"]:
                reasons.append("Linux memory full PSI exceeded 1% between samples")
        except KeyError:
            unknown.append("Linux memory PSI unavailable")
    if samples[-1]["swapout_pages"] - samples[0]["swapout_pages"] > thresholds["swapout_pages"]:
        reasons.append("host swapouts increased by more than 1024 pages")
    status = "contended" if reasons else ("unverified_environment" if unknown else "valid")
    return {"status": status, "reasons": sorted(set(reasons + unknown)), "thresholds": thresholds,
            "sample_count": len(samples),
            "limitations": "Guest OS counters cannot establish that a shared runner's physical host is idle."}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("operation", choices=["snapshot", "assess"])
    parser.add_argument("samples", nargs="?")
    args = parser.parse_args()
    if args.operation == "snapshot":
        print(json.dumps(snapshot()))
    else:
        with open(args.samples) as stream:
            print(json.dumps(assess([json.loads(line) for line in stream if line.strip()])))
