import http from 'k6/http';

export let options = {
    vus: 300,
    duration: '30s',
};

export default function () {
    const url = 'http://localhost:8080/flashsale/precheck';

    const uid = Math.floor(Math.random() * 2000000);

    http.post(url, JSON.stringify({
        user_id: `bot_${uid}`,
        product: "p1",
    }), {
        headers: {
            "Content-Type": "application/json",
            "User-Agent": `Bot-${uid}`
        }
    });
}
