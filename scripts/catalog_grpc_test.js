import grpc from 'k6/net/grpc';
import { check } from 'k6';

const client = new grpc.Client();
client.load(['proto/catalog/v1', '../proto/catalog/v1'], 'catalog.proto');

// Skenario pengujian performa gRPC Catalog Service (Downstream)
export const options = {
  vus: 50,
  duration: '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default () => {
  client.connect('localhost:50052', { plaintext: true });
  const payload = {
    product_id: 'p0000001-0000-0000-0000-000000000001',
  };
  const response = client.invoke('supermart.catalog.v1.CatalogService/GetProduct', payload);
  check(response, {
    'status is OK': (r) => r !== null && r.status === grpc.StatusOK,
    'has product data': (r) => r !== null && r.message && r.message.product && r.message.product.id !== '',
  });
  client.close();
};

