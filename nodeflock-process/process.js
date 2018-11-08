#!/usr/bin/env node
var ___nodejsrequire = require;
var fs = ___nodejsrequire('fs');
var ConsoleRedirect = /** @class */ (function () {
    function ConsoleRedirect(enabled, logFile, console, logBuffer) {
        if (enabled === void 0) { enabled = true; }
        if (logBuffer === void 0) { logBuffer = []; }
        this.enabled = enabled;
        this.logFile = logFile;
        this.console = console;
        this.logBuffer = logBuffer;
        var consoleRedirect = this;
        if (this.enabled) {
            this.fileFuncs = ConsoleRedirect.makeFuncs(function (level) {
                return ConsoleRedirect.getLogFunc(level, consoleRedirect.logToFile, consoleRedirect);
            });
            this.collectFuns = ConsoleRedirect.makeFuncs(function (level) {
                return ConsoleRedirect.getLogFunc(level, consoleRedirect.appendToBuffer, consoleRedirect);
            });
        }
        else {
            this.fileFuncs = ConsoleRedirect.makeFuncs(function (level) {
                return ConsoleRedirect.getNullFunc();
            });
            this.collectFuns = ConsoleRedirect.makeFuncs(function (level) {
                return ConsoleRedirect.getNullFunc();
            });
        }
        this.redirectToFile();
    }
    ConsoleRedirect.prototype.logToFile = function (level) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        if (args.length == 0) {
            return;
        }
        var d = (new Date);
        var lines = [];
        for (var i in args) {
            var arg = args[i];
            if ("function" == typeof arg) {
                arg = "[Function]";
            }
            lines.push(d + " level " + level + " " + JSON.stringify(arg) + "\n");
        }
        fs.appendFile(this.logFile, lines.join(""), function (err) {
            if (err) {
                return console.error("could not write to file", err);
            }
        });
    };
    ConsoleRedirect.prototype.appendToBuffer = function (level) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        this.logBuffer.push({
            level: level,
            data: args,
            stack: ConsoleRedirect.getStack()
        });
    };
    ConsoleRedirect.makeFuncs = function (logFuncConstructor) {
        var funcs = {};
        for (var i in ConsoleRedirect.consoleProps) {
            var level = ConsoleRedirect.consoleProps[i];
            funcs[level] = logFuncConstructor(level);
        }
        return funcs;
    };
    ConsoleRedirect.prototype.wire = function (funcs) {
        for (var i in ConsoleRedirect.consoleProps) {
            var level = ConsoleRedirect.consoleProps[i];
            this.console[level] = funcs[level];
        }
    };
    ConsoleRedirect.getNullFunc = function () {
        return function () { };
    };
    ConsoleRedirect.getLogFunc = function (level, f, consoleRedirect) {
        return function () {
            var a = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                a[_i] = arguments[_i];
            }
            f.apply(consoleRedirect, [level].concat(a));
        };
    };
    ConsoleRedirect.prototype.redirectToFile = function () {
        this.wire(this.fileFuncs);
    };
    ConsoleRedirect.prototype.flushBufferToFile = function () {
        for (var i in this.logBuffer) {
            var entry = this.logBuffer[i];
            this.logToFile.apply(this, [entry.level, entry.data]); //.concat(entry.stack));
        }
        this.logBuffer = [];
    };
    ConsoleRedirect.prototype.collect = function () {
        this.wire(this.collectFuns);
    };
    ConsoleRedirect.getStack = function () {
        var lines = new Error().stack.split("\n");
        if (lines.length > 4) {
            return lines.slice(4);
        }
        return lines;
    };
    ConsoleRedirect.prototype.flushBuffer = function () {
        var b = this.logBuffer;
        this.logBuffer = [];
        return b;
    };
    ConsoleRedirect.consoleProps = ["log", "info", "warn", "error"];
    return ConsoleRedirect;
}());
function resolveComponent(name, components) {
    if (name.length == 0) {
        return undefined;
    }
    if (name.length == 1) {
        if (components.hasOwnProperty(name[0])) {
            return components[name[0]];
        }
        return undefined;
    }
    if (components.hasOwnProperty(name[0])) {
        return resolveComponent(name.slice(1), components[name[0]]);
    }
    return undefined;
}
function executeCall(components, callBuffer, consoleRedirect, resultHandler) {
    var call;
    var error = "";
    var result = undefined;
    var async = false;
    // maybe a timeout?        
    try {
        // let us see if we can read the incoming JSON
        call = JSON.parse(callBuffer.toString());
        var componentFunc = resolveComponent(call.func.split("."), components);
        if (componentFunc === undefined) {
            error = "component func: \"" + call.func + "\" not found";
        }
        else {
            try {
                // let us start collecting console output before we call the hosted js
                consoleRedirect.collect();
                var callResult = componentFunc.apply(null, call.args);
                if ("function" != typeof callResult) {
                    // that was synchonous
                    result = callResult;
                }
                else {
                    // they returned a function to pass in a callback that 
                    // will process their result
                    try {
                        async = true;
                        callResult(function (endResult) {
                            resultHandler({
                                result: endResult,
                                error: error,
                                log: consoleRedirect.flushBuffer()
                            });
                            consoleRedirect.redirectToFile();
                        });
                    }
                    catch (asyncCallErr) {
                        // let us go back to non async
                        async = false;
                    }
                }
            }
            catch (e) {
                error = "error calling \"" + call.func + "\": " + e.message;
            }
            if (!async) {
                consoleRedirect.redirectToFile();
            }
        }
    }
    catch (e) {
        // json parsing failed
        error = "could not parse incoming json: " + e.message;
    }
    if (!async) {
        // synchonous "callback" of the resultHandler            
        resultHandler({
            result: result,
            error: error,
            log: consoleRedirect.flushBuffer()
        });
    }
}
function run(components, consoleRedirect) {
    var readBuffer = new Buffer("");
    process.stdin.on('readable', function () {
        var buffer = process.stdin.read(); // 38{"func": "foo.bar", "args": ["hallo"]}
        if (buffer) {
            readBuffer = Buffer.concat([readBuffer, buffer]);
        }
        // find start
        var callLength = 0;
        var callStart = 0;
        for (var i = 0; i < readBuffer.length; i++) {
            if (readBuffer[i] == 123) { // opening "{"
                callStart = i;
                callLength = parseInt(readBuffer.slice(0, callStart).toString());
                break;
            }
        }
        if (callLength > 0 && (readBuffer.length >= (callStart + callLength))) {
            // got a valid call
            executeCall(components, readBuffer.slice(callStart, callStart + callLength), consoleRedirect, function (result) {
                var resultBuffer = new Buffer(JSON.stringify(result));
                process.stdout.write(resultBuffer.length.toString());
                process.stdout.write(resultBuffer);
                readBuffer = readBuffer.slice(callStart + callLength);
            });
        }
    });
    process.stdin.on('end', function () {
        process.stdout.write('std in was closed with a readbuffer ' + readBuffer.toString());
        process.exit(1);
    });
}
var ____consoleRedirect = new ConsoleRedirect(true, "/tmp/test.log", console);
____consoleRedirect.collect();
var jsSource = process.argv.pop();
var contents = fs.readFileSync(jsSource).toString();
console.log("starting", process.pid);
process.on('uncaughtException', function (e) {
    ____consoleRedirect.logToFile("______________________________ uncaughtException", typeof e);
    ____consoleRedirect.logToFile("uncaught", e);
    console.log("uncaught", e);
    // application specific logging, throwing an error, or other logic here
});
eval("(function(require, module, ___nodejsrequire) {\n    " + contents + "\n})(undefined, undefined, ___nodejsrequire);");
____consoleRedirect.flushBufferToFile();
____consoleRedirect.redirectToFile();
run(global, ____consoleRedirect);
