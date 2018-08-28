var assert = require('assert');
var fw = require("..");

describe('Array', function () {
    describe('#indexOf()', function () {
        it('should return -1 when the value is not present', function () {
            assert.equal([1, 2, 3].indexOf(4), -1);
        });
    });
});

describe('Task', function() {
    describe('task()', function() {
        it('should return a valid TaskSpec', function () {
            fw.task({
                run: "fooFn",
                inputs: {
                    "message": "hello world!",
                }
            });
        });
    });

    describe('#require()', function() {
        it('should add single dependency to task when given a string', function () {
            var myTask = fw.task({
                run: "fooFn",
                inputs: {
                    "message": "hello world!",
                }
            });
            const depID = "someDependency";
            myTask.require(depID);
            assert.deepEqual(myTask.requires[depID], {});
        })
        it('should add all dependencies to task when given an array', function () {
            var myTask = fw.task({
                run: "fooFn",
                inputs: {
                    "message": "hello world!",
                }
            });
            const depIDs = ["A", "B", "C"];
            myTask.require(depIDs);
            assert.deepEqual(Object.keys(myTask.requires), depIDs);
        })
    });
});