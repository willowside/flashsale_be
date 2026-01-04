import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '5s', target: 50 },
    { duration: '20s', target: 200 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  const userId = __VU % 10 === 0 ? 'force-fail' : `user-${__VU}-${__ITER}`;
  const productId = '1001';

  // Precheck
  const res = http.post(
    `${BASE_URL}/flashsale/precheck`,
    JSON.stringify({
      user_id: userId,
      product_id: productId,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(res, {
    'precheck accepted': r => r.status === 200,
  });

  sleep(0.1);
}
