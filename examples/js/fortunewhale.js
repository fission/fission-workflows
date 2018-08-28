var fw = require("fission-workflows");

module.exports = function () {
    return fw.sequence([
        fw.task({
            id: "generateFortune", // optional ID
            run: "fortune"
        }),
        fw.sleep(1000),
        fw.task({
            run: "whalesay",
            inputs: {
                "body": "{ output('generateFortune') }"
            }
        })
    ])
};
