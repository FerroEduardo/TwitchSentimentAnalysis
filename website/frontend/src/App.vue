<script setup>
import { computed, onMounted, ref } from 'vue';
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale,
  LineElement,
  PointElement,
  Filler
} from 'chart.js'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { parseChartData } from './helper/chart';
import { connect } from './helper/websocket';

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement, Filler)

const results = ref([])
const messages = ref([])

const charts = computed(() => {
  const charts = {}

  Object.entries(results.value)
    .forEach(
      ([channelName, channelData]) => {
        charts[channelName] = parseChartData(channelName, channelData);
      }
    );

  return charts
})

function formatSentimentNumber(value) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
}

onMounted(() => connect(
  (event) => {
    const payload = JSON.parse(event.data)

    switch (payload.event) {
      case "results":
        results.value = payload.data
        break;
      case "messages":
        messages.value = payload.data
        break;
    }
  }
))
</script>

<template>
  <div class="flex h-screen" v-if="Object.keys(charts) == 0">
    <div class="m-auto border rounded p-8">
      No data available
    </div>
  </div>
  <div v-else class="max-w-[1200px] flex flex-col gap-4 lg:mx-auto lg:my-8 m-2">
    <Card v-for="(chart, channel) in charts" :key="channel" class="w-full">
      <CardHeader>
        <CardTitle>{{ channel }}</CardTitle>
      </CardHeader>
      <CardContent class="lg:max-h-96 max-h-[600px] flex flex-col lg:flex-row gap-1">
        <Line :data="chart.data" :options="chart.options" class="lg:max-w-[600px] chart max-w-full" style="max-height: 384px;"/>
        <Table class="h-20">
          <TableHeader>
            <TableRow>
              <TableHead class="lg:w-36">User</TableHead>
              <TableHead>Message</TableHead>
              <TableHead class="lg:w-36 text-center">Sentiment</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody class="h-full">
            <TableRow v-if="(messages[channel] ?? []).length == 0">
              <TableCell colspan="3" class="text-center">
                No data available 
              </TableCell>
            </TableRow>
            <TableRow v-else v-for="(message) in (messages[channel] ?? []).slice(0, 10)" :key="message.id">
              <TableCell>{{ message.user }}</TableCell>
              <TableCell>{{ message.message }}</TableCell>
              <TableCell class="text-center">
                <span style="color: rgba(75, 192, 192, 1);">{{ formatSentimentNumber(message.sentiment_positive)}}</span>/<span style="color: rgba(255, 206, 86, 1);">{{ formatSentimentNumber(message.sentiment_neutral)}}</span>/<span style="color: rgba(255, 99, 132, 1);">{{ formatSentimentNumber(message.sentiment_negative)
                  }}</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>

<style></style>
