import grpc from 'k6/net/grpc';
import { check } from 'k6';

const client = new grpc.Client();
client.load(['../proto/inventory/v1'], 'inventory.proto');

// Skenario 1: 50 VU selama 30 detik
export const options = { vus: 50, duration: '30s', summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'] };

export default () => {
  client.connect('localhost:50051', { plaintext: true });
  const payload = {
    product_id: 'p0000001-0000-0000-0000-000000000001',
    warehouse_id: 'w0000001-0000-0000-0000-000000000001'
  };
  const response = client.invoke('supermart.inventory.v1.InventoryService/GetStockItem', payload);
  check(response, { 'status is OK': (r) => r !== null && r.status === grpc.StatusOK });
  client.close();
};