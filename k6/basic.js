import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    vus: 10,
    duration: '30s',
};

export default function () {
    const url = 'http://localhost:8080/flashsale/precheck';

    let payload = JSON.stringify({
        user_id: `user_${__VU}_${__ITER}`,
        product: 'p1',
    });

    let headers = { 'Content-Type': 'application/json' };

    http.post(url, payload, { headers });

    sleep(0.2);
}
