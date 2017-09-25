'use strict';
//
// Wunderground Get
//
// Gets the current weather for a specific city.
//
// Input:
//{
//  city: <string>
//  apiKey: <string
//}
//
// Output:
// See example
//
// Example
// ---------
// Request:
// {
//   city: San Fransisco
//   apiKey: <your_key>
// }

let https = require('https');

async function fetchWeather(apiKey, state, city) {
    const url = `https://api.wunderground.com/api/${apiKey}/conditions/q/${state}/${city}.json`;
    console.log(`GET ${url}`);
    return new Promise(function (resolve, reject) {
        let req = https.request(url, function (res) {

            let rawData = '';
            res.on('data', (chunk) => {
                rawData += chunk;
            });
            res.on('end', () => {
                try {
                    if (res.statusCode >= 400) {
                        reject(res.statusCode);
                        return
                    }
                    const parsedData = JSON.parse(rawData);
                    resolve(parsedData);
                }
                catch (e) {
                    console.error(e.message);
                }
            });
        });
        req.end();
    });
}

module.exports = async function (context) {
    const b = context.request.body;
    if (!b) {
        return {
            status: 400,
            body: 'missing body',
        };
    }
    if (!b.apiKey) {
        return {
            status: 400,
            body: 'missing Wunderground API key',
        };
    }
    if (!b.state) {
        return {
            status: 400,
            body: 'missing location state',
        };
    }
    if (!b.city) {
        return {
            status: 400,
            body: 'missing location city',
        };
    }
    const apiKey = b.apiKey;
    const state = b.state;
    const city = b.city;
    try {
        const resp = await fetchWeather(apiKey, state, city);
        console.log(`RESULT ${resp}`);
        return {
            status: 200,
            body: resp,
            headers: {
                'Content-Type': 'application/json'
            }
        }
    } catch (e) {
        return {
            status: 500,
            body: e.message
        }
    }
};
