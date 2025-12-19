import http from 'k6/http';

export let options = {
    vus: 150,
    duration: '20s',
};

export default function () {
    const url = 'http://localhost:8080/flashsale/precheck';

    let headers = {
        "Content-Type": "application/json",
        "X-User-ID": `spoof_${Math.floor(Math.random() * 999999)}`,
        "User-Agent": "FakeBrowser/9.9"
    };

    http.post(url, JSON.stringify({ product_id: "1001" }), { headers });
}
