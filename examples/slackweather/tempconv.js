'use strict';
//
// tempconv
//
// A function that converts a temperature between formats
//
// Usage:
// {
//   temperature: <float|int>
//   format: <C|F|K>
//   target: <C|F|K>
// }
//
// Output:
// {
//   temperature: <float|int>
//   target: <C|F|K>
// }

// Formats is a 2d-map, in which the outer dimension is the source and the inner dimension is the target format.
const conversions = {
    "C": {
        "K": (temp) => {
            return temp + 273.15
        },
        "F": (temp) => {
            return temp * (9 / 5) + 32
        }
    },
    "K": {
        "C": (temp) => {
            return temp - 273.15
        },
        "F": (temp) => {
            return temp * (9 / 5) - 459.67
        }
    },
    "F": {
        "C": (temp) => {
            return (temp - 32) * (5 / 9)
        },
        "K": (temp) => {
            return (temp + 459.67) * (5 / 9)
        },
    },
};

function convert(temp, format, target) {
    if (format === target) {
        return temp
    }
    const src = conversions[format];
    if (!src) {
        throw new Error(`unknown temperature format '${format}'`)
    }
    const conv = src[target];
    if (!conv) {
        throw new Error(`unknown temperature target '${target}'`)
    }

    return conv(temp)
}

module.exports = async function (context) {
    const b = context.request.body;
    console.log("body", b);
    if (!b) {
        return {
            status: 400,
            body: 'missing body',
        };
    }
    if (!b.temperature) {
        return {
            status: 400,
            body: 'missing temperature',
        };
    }
    if (!b.format) {
        return {
            status: 400,
            body: 'missing temperature format',
        };
    }
    if (!b.target) {
        return {
            status: 400,
            body: 'missing temperature target',
        };
    }
    const temperature = parseFloat(b.temperature);
    const format = b.format.toUpperCase().trim();
    const target = b.target.toUpperCase().trim();
    try {
        const converted = convert(temperature, format, target);
        console.log(`result: ${converted}`);
        return {
            status: 200,
            body: {
                temperature: Math.round(converted),
                format: target
            },
            headers: {
                'Content-Type': 'application/json'
            }
        }
    } catch (e) {
        return {
            status: 400,
            body: e.message
        }
    }
};
