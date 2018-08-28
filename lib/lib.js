var pb = require("./types.pb.js");
// TODO parse input values
// TODO validate tasks, workflows, inputs

// Extensions to generated protobuf types.

pb.fission.workflows.types.TaskSpec.prototype.require = function (tasks) {
    if (!Array.isArray(tasks)) {
        tasks = [tasks]
    }
    const _this = this;
    tasks.forEach(function (task) {
        _this.requires[task] = {}
    });
    return this;
};

// Primitives
const models = {
    workflow(outputTask, tasks) {
        return new pb.fission.workflows.types.WorkflowSpec({
            outputTask: outputTask,
            tasks: tasks,
        })
    },
    task(params) {
        // TODO parse function-expressions to strings
        // TODO generate ID if not set
        return new pb.fission.workflows.types.TaskSpec({
            id: params.id,
            run: params.run,
            inputs: params.inputs || {},
            requires: params.requires || {},
        })
    }
};

// Library
const stdlib = {
    sleep(milliseconds) {
        return models.task({
            run: internalFnref("sleep"),
            inputs: {
                "duration": milliseconds + "ms"
            }
        })
    },
    noop() {
        return models.task({
            run: internalFnref("noop"),
        })
    },
    compose(object) {
        return models.noop({
            run: internalFnref("compose"),
            inputs: object,
        })
    },
    http(params) {
        return models.task({
            run: internalFnref("http"),
            inputs: {
                "url": params.url,
                "headers": params.headers,
                "content-type": params.contentType,
                "body": params.body,
            }
        })
    },
    fail(msg) {
        return models.task({
            run: internalFnref("fail"),
            inputs: {
                "message": msg,
            }
        })
    },
    js(fn, inputs) {
        return models.task({
            run: internalFnref("javascript"),
            inputs: {
                fn: fn.toString(),
                args: inputs,
            }
        })
    },

    //
    // Dynamic Workflow functions
    //

    if(expr, then, orElse) {
        return models.task({
            run: internalFnref("if"),
            inputs: {
                expr: expr,
                then: then,
                orElse: orElse,
            },
        })
    },
    foreach(items, consumer, sequential, collect) {
        return models.task({
            run: internalFnref("foreach"),
            inputs: {
                foreach: items,
                do: consumer,
                sequential: sequential,
                collect: collect,
            },
        })
    },
    switch(val, cases, defaultValue) {
        return models.task({
            run: internalFnref("switch"),
            inputs: {
                switch: val,
                cases: cases,
                default: defaultValue,
            }
        })
    },
    while(condition, action, limit) {
        return models.task({
            run: internalFnref("while"),
            inputs: {
                expr: condition,
                do: action,
                limit: limit,
            }
        })
    },
};

// Control Flow constructs

const controlFlowConstructs = {
    sequence(tasks) {
        var lastTask = null;
        tasks.forEach(function (task) {
            if (lastTask != null) {
                task.require(lastTask)
            }
            lastTask = task
        });
        return lastTask
    }
};

// util

function deepCopy(obj) {
    return JSON.parse(JSON.stringify(obj))
}

function internalFnref(fnref) {
    return "internal://" + fnref
}

// Module config

Object.assign(module.exports, models, stdlib, controlFlowConstructs);