import { runBenchmarks } from './simulator';
import { infrastructure } from './infrastructure';

const latencyBenchmark = await runBenchmarks({
  clients: {
    [infrastructure.areas[0].host]: 60,
    [infrastructure.areas[0].areas[0].host]: 40,
    [infrastructure.areas[0].areas[1].host]: 20,
  },
  requests: 1000,
  requestsPerSecond: 10,
});

const unbalancedClientsDistribution = await runBenchmarks({
  clients: {
    [infrastructure.areas[0].host]: 100,
  },
  requests: 1000,
  requestsPerSecond: 10,
});

// Output results.
console.log('Latency benchmark results:', JSON.stringify(latencyBenchmark));
console.log('Unbalanced clients distribution results:', JSON.stringify(unbalancedClientsDistribution));
