import http from 'k6/http';
import exec from 'k6/execution';
import { Counter, Rate, Trend } from 'k6/metrics';

const successes = new Counter('business_successes');
const successLatency = new Trend('success_latency', true);
const failureLatency = new Trend('failure_latency', true);
const errors = new Rate('business_errors');
const badEvents = new Rate('slo_bad_events');
const rate = Number(__ENV.RATE || 5);
const seconds = Number(__ENV.WORKLOAD_SECONDS || 60);
const slo = Number(__ENV.SLO_MS || 500);
const dynamic = __ENV.PROFILE === 'step';
// Requests time out after nine seconds, so this covers the worst case.
const preAllocatedVUs = Number(__ENV.PREALLOCATED_VUS || 20);
const maxVUs = Number(__ENV.MAX_VUS || Math.max(preAllocatedVUs, rate * 15));
if (!Number.isInteger(preAllocatedVUs) || preAllocatedVUs < 1 ||
    !Number.isInteger(maxVUs) || maxVUs < preAllocatedVUs) {
  throw new Error('VU limits must be positive integers and max must cover preallocation');
}
const base = __ENV.BASE_URL || 'http://127.0.0.1:8080';
if (!/^http:\/\/(127\.0\.0\.1|localhost):\d+$/.test(base)) {
  throw new Error('This workload only targets a local lab endpoint');
}

export const options = {
  scenarios: {
    orders: dynamic ? {
      executor: 'ramping-arrival-rate',
      startRate: Math.max(1, Math.floor(rate / 4)),
      timeUnit: '1s',
      preAllocatedVUs,
      maxVUs,
      stages: [
        { target: Math.max(1, Math.floor(rate / 4)), duration: `${Math.floor(seconds / 4)}s` },
        { target: rate, duration: '1s' },
        { target: rate, duration: `${Math.floor(seconds / 2)}s` },
        { target: Math.max(1, Math.floor(rate / 4)), duration: '1s' },
        { target: Math.max(1, Math.floor(rate / 4)), duration: `${Math.max(1, seconds - Math.floor(seconds / 4) - Math.floor(seconds / 2) - 2)}s` },
      ],
      gracefulStop: '10s',
    } : {
      executor: 'constant-arrival-rate',
      rate,
      timeUnit: '1s',
      duration: `${seconds}s`,
      preAllocatedVUs,
      maxVUs,
      gracefulStop: '10s',
    },
  },
  summaryTrendStats: ['avg', 'min', 'max', 'p(50)', 'p(95)', 'p(99)'],
  discardResponseBodies: false,
  // kube-proxy balances per connection, so reused connections skew replica load.
  noConnectionReuse: __ENV.CONNECTION_REUSE !== 'true',
};

export default function () {
  const customer = 1 + (exec.scenario.iterationInTest % 10000);
  const response = http.post(`${base}/orders`, JSON.stringify({
    customer_id: customer,
    items: [{ price: 1000, quantity: 2 }],
  }), { headers: { 'Content-Type': 'application/json' }, timeout: '9s', tags: { name: 'POST /orders' } });
  let ok = false;
  try { ok = response.status === 201 && response.json('status') === 'paid'; } catch (_) { ok = false; }
  successes.add(ok ? 1 : 0);
  errors.add(!ok);
  badEvents.add(!ok || response.timings.duration > slo);
  if (ok) successLatency.add(response.timings.duration);
  else failureLatency.add(response.timings.duration);
}

export function handleSummary(data) {
  const m = data.metrics;
  const value = (name, field) => m[name] && m[name].values[field] !== undefined ? m[name].values[field] : null;
  const summary = {
    duration_seconds: data.state.testRunDurationMs / 1000,
    requests: value('http_reqs', 'count'),
    successes: value('business_successes', 'count'),
    error_rate: value('business_errors', 'rate'),
    slo_bad_rate: value('slo_bad_events', 'rate'),
    success_p95_ms: value('success_latency', 'p(95)'),
    success_p99_ms: value('success_latency', 'p(99)'),
    failure_p95_ms: value('failure_latency', 'p(95)'),
    dropped_iterations: value('dropped_iterations', 'count') || 0,
    raw: data,
  };
  return { [__ENV.SUMMARY_PATH || 'results/k6-summary.json']: JSON.stringify(summary, null, 2) };
}
