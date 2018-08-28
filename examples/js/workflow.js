import fw from "fission/workflows";


// Basic functionality
const fetchedWeather = fw.task("weather-fetch");
fetchedWeather.inputs["location"] = "Delft, The Netherlands";

const waitSecond = fw.sleep(1000);
waitSecond.requires.push(fetchedWeather);

exports.flow = fw.task({
    name: "customName",
    run: "slack-send",
    inputs: {
        "msg": "<expr>",
    },
    requires: [
        waitSecond,
    ]
});

// Using control flow constructs
exports.flow = fw.sequence(
    fw.task({
        run: "weather-fetch",
        inputs: {
            "location": "Delft, The Netherlands"
        }
    }),
    fw.sleep(1000),
    fw.task({
        run: "slack-send",
        inputs: {
            "msg": "<expr>"
        }
    }),
);

//
// Ideas for expressions
//
// 1. original
fw.expr("$.Tasks.weather-fetch.Output");

// 2. specific helpers
fw.expr.output("taskA");
fw.expr.input("taskA", "key"); // 1 argument == param

// 3. 'promise'
// - Downside
myTask.promise().output();
myTask.promise().input("key");

// 4. Function
// - serialize function to string, unpack in Otto and evaluate with scope.
// - requires
inputs["myFunc"] = (workflow) => { return workflow.tasks["weather-fetch"].output };

//
// Idea for tasks
//
// Allow function instead of function reference
// - This will be run in a javascript function (needs to be adapted)
fw.task({
    run: function(inputs) {
        return "myOutput"
    },
    inputs: {
        // ...
    }
});