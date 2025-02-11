import pandas as pd
import matplotlib.pyplot as plt

# Read the CSV files
edge_only_df = pd.read_csv('./edge_only.csv')
mixed_df = pd.read_csv('./mixed.csv')

# Plot the data
plt.figure(figsize=(10, 6))
plt.plot(edge_only_df['Requests'], edge_only_df['AvgResponseTime'], label='Edge Only')
plt.plot(mixed_df['Requests'], mixed_df['AvgResponseTime'], label='Edge + Cloud')
plt.xlabel('Number of Parallel Requests')
plt.ylabel('Average Response Time (ms)')
plt.title('Average Response Time vs. Number of Parallel Requests')
plt.ylim(0, None)
plt.legend()
plt.grid(True)
plt.show()
