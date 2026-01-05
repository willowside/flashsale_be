import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '5s', target: 50 },   // 爬升階段
    { duration: '20s', target: 200 },  // 穩定壓力階段 (200 VUs 對單一 Worker 已有足夠壓力)
    { duration: '10s', target: 0 },    // 冷卻階段
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],  // Gatekeeper 模式下，API 響應應非常快
    http_req_failed: ['rate<0.15'],   // 預期會有 409 (售罄/重複) 與 429 (限流)，故放寬門檻
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  // 1. 模擬重複購買：約 15% 的請求使用固定 ID
  const isDuplicate = __ITER % 7 === 0;
  const userId = isDuplicate ? `user-${__VU}-fixed` : `user-${__VU}-${__ITER}`;

  // 2. 模擬系統失敗：約 2% 的請求觸發 force-fail 以驗證 DLQ 補償
  const isForceFail = __ITER % 50 === 0;
  const finalUserId = isForceFail ? 'force-fail' : userId;

  const payload = JSON.stringify({
    user_id: finalUserId,
    product_id: '1001',
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(`${BASE_URL}/flashsale/precheck`, payload, params);

  // 斷言邏輯
  check(res, {
    'is status 200 or 409 or 429': (r) => [200, 409, 429].includes(r.status),
    'precheck accepted': (r) => r.status === 200,
    'already purchased': (r) => r.status === 409 && r.body.includes("USER_ALREADY_PURCHASED"),
  });

  // 模擬真實使用者思考時間 (50ms - 100ms)
  sleep(Math.random() * 0.05 + 0.05);
}