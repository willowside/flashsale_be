import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '5s', target: 100 },  // 快速爬升
    { duration: '20s', target: 300 }, // 衝擊 300 VUs，觀察多個 Worker 的處理能力
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 門衛模式應保持極低延遲
    http_req_failed: ['rate<0.1'],   // 預期 409/429 會發生，因此放寬門檻
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  // 模擬 15% 的用戶會重複發送請求，驗證 precheck_final.lua 的 SADD/SISMEMBER 是否生效
  const isDuplicate = __ITER % 7 === 0; 
  const userId = isDuplicate ? `user-${__VU}-fixed` : `user-${__VU}-${__ITER}`;
  
  // 隨機測試 force-fail 以觸發補償機制
  const isForceFail = __ITER % 50 === 0;
  const finalUserId = isForceFail ? 'force-fail' : userId;

  const res = http.post(
    `${BASE_URL}/flashsale/precheck`,
    JSON.stringify({
      user_id: finalUserId,
      product_id: '1001',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(res, {
    'accepted (200)': r => r.status === 200,
    'already purchased (409)': r => r.status === 409,
    'out of stock (409)': r => r.status === 409, // 方案 A 中售罄通常也是 409
    'rate limited (429)': r => r.status === 429,
  });

  sleep(0.05); // 縮短間隔以模擬更高頻率的搶購
}