import http from 'k6/http';

export let options = {
    vus: 200,
    duration: '30s',
};

export default function () {
    const url = 'http://localhost:8080/flashsale/precheck';

    const payload = JSON.stringify({
        user_id: "attacker_1",
        product: "p1",
    });

    http.post(url, payload, { headers: { "Content-Type": "application/json" } });
}
