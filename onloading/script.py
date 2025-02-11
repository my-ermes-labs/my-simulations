import pandas as pd
import matplotlib.pyplot as plt

# Read data from CSV
data = pd.read_csv('responses.csv')

# Categorize the responses
response_200 = data[data['ResponseCode'] == 200]
response_503 = data[data['ResponseCode'] == 503]
response_redirect = data[(data['ResponseCode'] == 301) | (data['ResponseCode'] == 302)]

# Plot the data
plt.figure(figsize=(10, 6))
plt.bar(response_200['Time'], response_200['ResponseTime'], color='#00FF00', edgecolor='black', label='200 response')  # Bright green
plt.bar(response_503['Time'], response_503['ResponseTime'], color='#FF6347', edgecolor='black', label='503 response')  # Tomato red
plt.bar(response_redirect['Time'], response_redirect['ResponseTime'], color='#FFD700', edgecolor='black', label='redirect + 200 response')  # Gold

plt.xlabel('Time (ms)')
plt.ylabel('Response Time (ms)')
plt.title('Client Requests and Response Times with Session Migration')
plt.legend()
plt.grid(True)
plt.show()
