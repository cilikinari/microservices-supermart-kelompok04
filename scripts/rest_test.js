import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 50,
  duration: '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

const baseURL = __ENV.BASE_URL || 'http://localhost:3000';
const email = __ENV.AUTH_EMAIL || 'gudang.jakarta@nusantara.co.id';
const password = __ENV.AUTH_PASSWORD || 'password123';

export function setup() {
  const response = http.post(
    `${baseURL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const loginOK = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login token exists': (r) => Boolean(r.json('data.token')),
  });

  if (!loginOK) {
    throw new Error(`Login failed with status ${response.status}: ${response.body}`);
  }

  return { token: response.json('data.token') };
}

export default function (data) {
  const url =
    `${baseURL}/api/v1/inventory/stocks?warehouse_id=` +
    'w0000001-0000-0000-0000-000000000001';

  const response = http.get(url, {
    headers: { Authorization: 'Bearer ' + data.token },
  });

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
