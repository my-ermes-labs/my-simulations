import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

# Constants
TIME_SLICE_MS = 400  # time slice in milliseconds, can be edited as needed
CSV_FILE = 'response_times.csv'

def plot_graph():
    df = pd.read_csv(CSV_FILE)

    # Ensure the dataframe has the correct columns
    assert 'Only200ResponseTime' in df.columns and 'Non200ResponseTime' in df.columns, "CSV must have 'Only200ResponseTime' and 'Non200ResponseTime' columns"

    # Calculate the start times
    df['Start_Time'] = np.arange(0, len(df) * TIME_SLICE_MS, TIME_SLICE_MS)

    # Calculate total response times
    df['Total_Response'] = df['Only200ResponseTime'] + df['Non200ResponseTime']

    # Increase the figure size for better readability
    plt.figure(figsize=(14, 8))

    bar_width = TIME_SLICE_MS / 1000  # converting ms to seconds for the width of the bars

    # First value bars
    plt.bar(df['Start_Time']/1000, df['Only200ResponseTime'], width=bar_width, color='blue', edgecolor='black', linewidth=0.5, align='edge', label='Requests in Process')

    # Second value bars
    plt.bar(df['Start_Time']/1000, df['Non200ResponseTime'], width=bar_width, bottom=df['Only200ResponseTime'], color='orange', edgecolor='black', linewidth=0.5, align='edge', label='Requests being Migrated')

    # Set labels and title
    plt.xlabel('Time (seconds)')
    plt.ylabel('Average Response Time (ms)')
    plt.legend()

    # Increase the number of points on the x-axis and rotate labels
    plt.xticks(np.arange(0, (df['Start_Time'].max() + TIME_SLICE_MS) / 1000, TIME_SLICE_MS / 1000), rotation=45, ha='right')

    # Set the x-axis limits to avoid gaps
    plt.xlim(left=0, right=(df['Start_Time'].max() + TIME_SLICE_MS) / 1000)
    # plt.ylim(top=950)

    # Show the plot
    plt.show()

# Generate the graph
plot_graph()
