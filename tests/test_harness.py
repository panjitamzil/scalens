import copy
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("health", ROOT / "scripts/host-health.py")
health = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(health)


class HostHealthTests(unittest.TestCase):
    def samples(self):
        first = {"timestamp": 100, "os": "Linux", "boot_id": "one", "swapout_pages": 10000,
                 "cpu_ticks": [100, 0, 100, 800, 0, 0, 0, 0],
                 "memory_total_bytes": 1000, "memory_available_bytes": 500,
                 "pressure": {"memory": {"full": {"total": 1000}}}}
        second = copy.deepcopy(first)
        second.update(timestamp=110, cpu_ticks=[150, 0, 150, 1700, 0, 0, 0, 0])
        return [first, second]

    def test_existing_swap_is_not_contention(self):
        self.assertEqual(health.assess(self.samples())["status"], "valid")

    def test_swap_cpu_steal_and_pressure(self):
        changes = [
            {"swapout_pages": 12000},
            {"cpu_ticks": [1000, 0, 1100, 810, 0, 0, 0, 0]},
            {"cpu_ticks": [150, 0, 150, 1700, 0, 0, 0, 500]},
            {"memory_available_bytes": 10},
            {"pressure": {"memory": {"full": {"total": 501000}}}},
        ]
        for change in changes:
            with self.subTest(change=change):
                samples = self.samples()
                samples[-1].update(change)
                self.assertEqual(health.assess(samples)["status"], "contended")

    def test_missing_or_reset_data_never_valid(self):
        for change in [{"error": "no procfs"}, {"pressure": {}}, {"boot_id": "two"}, {"swapout_pages": 0}]:
            with self.subTest(change=change):
                samples = self.samples()
                samples[-1].update(change)
                self.assertEqual(health.assess(samples)["status"], "unverified_environment")
        self.assertEqual(health.assess([])["status"], "unverified_environment")

    def test_linux_parser(self):
        with tempfile.TemporaryDirectory() as tmp:
            proc = Path(tmp)
            (proc / "pressure").mkdir()
            (proc / "sys/kernel/random").mkdir(parents=True)
            (proc / "sys/kernel/random/boot_id").write_text("test\n")
            (proc / "meminfo").write_text("MemTotal: 2000 kB\nMemAvailable: 1000 kB\nSwapTotal: 100 kB\nSwapFree: 80 kB\n")
            (proc / "vmstat").write_text("pswpout 32\n")
            (proc / "stat").write_text("cpu 1 2 3 4 5 6 7 8 9 10\n")
            (proc / "pressure/memory").write_text("full avg10=0.00 avg60=0.00 avg300=0.00 total=100\n")
            result = health.linux_snapshot(proc)
            self.assertEqual(result["memory_available_bytes"], 1024000)
            self.assertEqual(result["swap_used_bytes"], 20480)
            self.assertEqual(result["cpu_ticks"], list(range(1, 9)))


class SuiteTests(unittest.TestCase):
    def test_aborted_run_saves_failure_and_calls_diagnostics(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "scripts").mkdir()
            (root / "bin").mkdir()
            shutil.copy(ROOT / "scripts/run-experiment.sh", root / "scripts")
            (root / "scripts/preflight.sh").write_text("exit 1\n")
            (root / "scripts/collect-diagnostics.sh").write_text('mkdir -p "$1"; touch "$1/called"\n')
            for name in ("dbys", "sleep"):
                path = root / "bin" / name
                path.write_text("#!/bin/sh\nexit 0\n")
                path.chmod(0o755)
            env = dict(os.environ, PATH=str(root / "bin") + os.pathsep + os.environ["PATH"],
                       RATE="2", SECONDS_PER_RUN="45", WARMUP_SECONDS="15", SEED_ROWS="1000",
                       CPU_WORK="1000", SCALE_REPLICAS="2")
            result = subprocess.run(["bash", str(root / "scripts/run-experiment.sh"), "cpu", "control", "1", "out"],
                                    env=env, text=True, capture_output=True)
            self.assertEqual(result.returncode, 1, result.stderr)
            failures = list(root.glob("out/control/*/failure.json"))
            self.assertEqual(len(failures), 1)
            self.assertEqual(json.loads(failures[0].read_text())["status"], "aborted")
            self.assertTrue((failures[0].parent / "diagnostics/called").exists())
            self.assertFalse((failures[0].parent / "run.json").exists())

    def suite(self, overrides=None, fail=False, mode="smoke"):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "scripts").mkdir()
            shutil.copy(ROOT / "scripts/run-suite.sh", root / "scripts")
            shutil.copytree(ROOT / "scenarios", root / "scenarios")
            (root / "scripts/run-experiment.sh").write_text(
                'printf "%s %s %s %s\\n" "$2" "$RATE" "$CPU_WORK" "$EXPERIMENT_MODE"\n'
                + ("exit 1\n" if fail else "exit 0\n"))
            env = {k: v for k, v in os.environ.items() if k not in
                   ("RATE", "CPU_WORK", "SEED_ROWS", "SECONDS_PER_RUN", "WARMUP_SECONDS", "REPEATS")}
            env.update(overrides or {})
            return subprocess.run(["bash", str(root / "scripts/run-suite.sh"), "cpu", "unused", mode],
                                  env=env, text=True, capture_output=True)

    def test_smoke_defaults_and_explicit_override(self):
        result = self.suite()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.splitlines(), ["control 2 1000 smoke", "scale 2 1000 smoke"])
        result = self.suite({"RATE": "3", "CPU_WORK": "0", "REPEATS": "2"})
        self.assertEqual(len(result.stdout.splitlines()), 4)
        self.assertIn("control 3 0 smoke", result.stdout)

    def test_failure_is_propagated_after_collecting_other_arms(self):
        result = self.suite(fail=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(len(result.stdout.splitlines()), 2)

    def test_calibration_is_one_repeat(self):
        result = self.suite(mode="calibration")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(len(result.stdout.splitlines()), 2)


if __name__ == "__main__":
    unittest.main()
