export const timestampFormat = new Intl.DateTimeFormat(
  undefined,
  {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: "h24"
  }
)

export function parseChartData(channelName, channelData) {
  const timestamps = channelData.map(item => timestampFormat.format(new Date(item.timestamp)));
  const avgSentimentPositive = channelData.map(item => item.avg_sentiment_positive);
  const avgSentimentNeutral = channelData.map(item => item.avg_sentiment_neutral);
  const avgSentimentNegative = channelData.map(item => item.avg_sentiment_negative);

  return {
    type: 'line',
    data: {
      labels: timestamps,
      datasets: [
        {
          label: 'Positive Sentiment',
          data: avgSentimentPositive,
          borderColor: 'rgba(75, 192, 192, 1)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
          borderWidth: 1,
          fill: true,
        },
        {
          label: 'Neutral Sentiment',
          data: avgSentimentNeutral,
          borderColor: 'rgba(255, 206, 86, 1)',
          backgroundColor: 'rgba(255, 206, 86, 0.2)',
          borderWidth: 1,
          fill: true,
        },
        {
          label: 'Negative Sentiment',
          data: avgSentimentNegative,
          borderColor: 'rgba(255, 99, 132, 1)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
          borderWidth: 1,
          fill: true,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        tooltip: {
          mode: 'index',
          intersect: false
        },
        hover: {
          mode: 'index',
          intersect: false
        },
        title: {
          display: false,
          text: `Channel: ${channelName}`,
        },
      },
      scales: {
        x: {
          title: {
            display: true,
            text: 'Timestamp',
          }
        },
        y: {
          title: {
            display: true,
            text: 'Average Sentiment',
          },
          beginAtZero: true,
          max: 1,
        },
      },
    },
  }
}