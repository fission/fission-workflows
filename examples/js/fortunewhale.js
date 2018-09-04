const fw = require("fission-workflows");

// Single-task workflow
module.exports = function () {
    return fw.task({
        run: "fortune"
    })
};

// Simple dependencies with tasks in the requires
module.exports = function () {
    const generateFortune = fw.Task({
        id: "generateFortune",
        run: "fortune"
    });

    const whaleSays = fw.task({
        run: "whalesay",
        inputs: {
            "body": "{ output('generateFortune') }"
        },
        requires: [
            generateFortune
        ]
    });

    return fw.Workflow({
        output: generateFortune,
        tasks: [
            generateFortune,
            whaleSays
        ]
    })
};

// Simple dependencies with task IDs in the requires
module.exports = function () {
    const generateFortune = fw.Task({
        id: "generateFortune",
        run: "fortune"
    });

    const whaleSays = fw.task({
        run: "whalesay",
        inputs: {
            "body": "{ output('generateFortune') }"
        },
        requires: [
            generateFortune.id
        ]
    });

    return fw.Workflow({
        output: generateFortune,
        tasks: [
            generateFortune,
            whaleSays
        ]
    })
};

// Referencing a task in an expression without an ID.
module.exports = function () {
    const generateFortune = fw.Task({
        run: "fortune"
    });

    const whaleSays = fw.task({
        run: "whalesay",
        inputs: {
            "body": `{ output('${generateFortune.id}') }`,  // or:
            "body2": output(generateFortune.id),            // or:
            "body3": output(generateFortune),
        },
        requires: [
            generateFortune.id
        ]
    });

    return fw.Workflow({
        output: generateFortune,
        tasks: [
            generateFortune,
            whaleSays
        ]
    })
};

// Sequence
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

// Expression using helper functions
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

// Expression with a function
module.exports = function () {
    return fw.sequence([
        fw.task({
            id: "generateFortune", // <- ID is optional, but needed to reference the function
            run: "fortune"
        }),
        fw.sleep(1000), // <- how to specify ID with these built-in functions?
        fw.task({
            run: "whalesay",
            inputs: {
                "body": function (scope) {
                    return scope.tasks["generateFortune"].output
                }
            }
        })
    ])
};

// Task with a function
module.exports = function () {
    return fw.sequence([
        fw.task({
            id: "generateFortune", // <- ID is optional, but needed to reference the function
            run: function (inputs) {
                return "Sometimes " + inputs["format"] + "is not the answer!"
            },
            inputs: {
                "format": "YAML"
            }
        }),
        fw.sleep(1000), // <- how to specify ID with these built-in functions?
        fw.task({
            run: "whalesay",
            inputs: {
                "body": "{ output('generateFortune') }"
            }
        })
    ])
};

/*
Discussion:
- Require IDs everywhere, or generate them if not provided (user will not be able to reference them)
- Expression functions are hard (impossible?) to validate beforehand; rely more on runtime errors.
 */