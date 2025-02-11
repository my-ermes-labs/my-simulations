import { ErmesClient } from '@ermes-labs/client';
import { performance } from 'perf_hooks';

export function runBenchmarks({ clients, requests, requestsPerSecond }: { clients: { [key: string]: number }, requests: number, requestsPerSecond: number }) {
  const clientInstances = Object.entries(clients).map(([host, count]) => new Array(count).fill(null).map(() => new ErmesClient({ initialOrigin: host })));
  const simulator = new Simulator(clientInstances.flat(), requests);

  return simulator.start(requestsPerSecond);
}

class Simulator {
  latencyRecords: number[] = [];
  completedRequests: number = 0;
  redirectionTime: number = 0;
  offloadingCount: number = 0;
  offloadedCount: number = 0;
  // Options.
  clients: ErmesClient[];
  totalRequests: number;

  constructor(clients: ErmesClient[], totalRequests: number) {
    this.clients = clients;
    this.totalRequests = totalRequests;
  }

  async makeRequest() {
    const clientNumber = Math.floor(Math.random() * this.clients.length);
    const startTime = performance.now();

    try {
      // We need to manually handle redirection to correctly make measurements.
      const response = await this.clients[clientNumber].fetch('/benchmark', { redirect: 'manual' });
      const endTime = performance.now();
      const latency = endTime - startTime;

      if (response.type === 'opaqueredirect') {
        this.completedRequests++;
        this.redirects++;

        const redirectUrl = response.headers.get('Location');
        console.log(`Request ${this.completedRequests + 1}: Redirect detected to ${redirectUrl}. Handling manually...`);
        this.handleRedirect(redirectUrl, startTime);
      }

      if (response.status === 200) {
        console.log(`Request ${this.completedRequests + 1}: OK in ${latency.toFixed(2)} ms`);
        this.latencyRecords.push(latency);
      } else if (response.status === 503) { // Assuming 503 for 'offloading'
        console.log(`Request ${this.completedRequests + 1}: Offloading, will retry.`);
        setTimeout(() => this.retryRequest(url, startTime), 100); // Retry after 100 ms
      } else if (response.type === 'opaqueredirect') {
        const redirectUrl = response.headers.get('Location');
        console.log(`Request ${this.completedRequests + 1}: Redirect detected to ${redirectUrl}. Handling manually...`);
        this.handleRedirect(redirectUrl, startTime);
      }

      this.completedRequests++;
    } catch (error) {
      console.error(`Error in request ${this.completedRequests + 1}: ${error}`);
    }
  }

  async retryRequest(url: string, originalStartTime: number) {
    const retryStartTime = performance.now();
    const response = await fetch(url, { redirect: 'manual' });
    const retryEndTime = performance.now();
    const totalRetryTime = retryEndTime - originalStartTime;

    this.offloadingCount++;
    this.redirectionTime += totalRetryTime;
    console.log(`Request retried: Completed in ${totalRetryTime.toFixed(2)} ms`);
  }

  async handleRedirect(url, originalStartTime) {
    const redirectStartTime = performance.now();
    const response = await fetch(url, { redirect: 'manual' });
    const redirectEndTime = performance.now();
    const totalRedirectTime = redirectEndTime - originalStartTime;

    this.offloadedCount++;
    this.redirectionTime += totalRedirectTime;
    console.log(`Redirect handled: Completed in ${totalRedirectTime.toFixed(2)} ms`);
  }

  start(rate: number) {
    return new Promise((resolve) => {
      const interval = 1000 / rate;
      const timer = setInterval(() => {
        if (this.completedRequests >= this.totalRequests) {
          clearInterval(timer);
          console.log("Simulation complete.");
          console.log(`Total Requests: ${this.totalRequests}, Offloading: ${this.offloadingCount}, Offloaded: ${this.offloadedCount}, Total Redirection Time: ${this.redirectionTime} ms`);
          resolve(this.latencyRecords);
        } else {
          this.makeRequest();
        }
      }, interval);
    });
  }
}

// Example usage:
const simulator = new RequestSimulator('https://example.com/api', 500, 10);
simulator.start(10); // Rate of 10 requests per second
