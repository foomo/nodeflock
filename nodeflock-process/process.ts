#!/usr/bin/env node
/// <reference path="node.d.ts" />
var fs = require('fs');

 
class ConsoleRedirect {
    private static consoleProps = ["log", "info", "warn", "error"];
    private fileFuncs;
    private collectFuns;
    constructor(
        private enabled:boolean = true,
        private logFile:string, 
        private console, 
        private logBuffer:any[] = []
    ) {
        var consoleRedirect = this;
        if(this.enabled) {
            this.fileFuncs = ConsoleRedirect.makeFuncs((level) => {
                return ConsoleRedirect.getLogFunc(level, consoleRedirect.logToFile, consoleRedirect);
            }) 
            this.collectFuns = ConsoleRedirect.makeFuncs((level) => {
                return ConsoleRedirect.getLogFunc(level, consoleRedirect.appendToBuffer, consoleRedirect);
            });
        } else {
            this.fileFuncs = ConsoleRedirect.makeFuncs((level) => {
                return ConsoleRedirect.getNullFunc();
            }) 
            this.collectFuns = ConsoleRedirect.makeFuncs((level) => {
                return ConsoleRedirect.getNullFunc();
            });
        }
        this.redirectToFile();
    }
    private logToFile(level, ...args) {
        for(var i in args) {
            fs.appendFile(this.logFile, (new Date) + " level " + level + " " + JSON.stringify(args[i]) + "\n", function (err) { 
                if(err) {
                    return console.error("could not write to file", err);
                }
            });
        }
    }
    private appendToBuffer(level, ...args) {
        this.logBuffer.push({
            level   : level,
            data : args,
            stack   : ConsoleRedirect.getStack()
        });
    }
    private static makeFuncs(logFuncConstructor) {
        var funcs = {};
        for(var i in ConsoleRedirect.consoleProps) {
            var level = ConsoleRedirect.consoleProps[i];
            funcs[level] = logFuncConstructor(level)
        }
        return funcs;
    }
    private wire(funcs) {
        for(var i in ConsoleRedirect.consoleProps) {
            var level = ConsoleRedirect.consoleProps[i];
            this.console[level] = funcs[level];
        }
    }
    private static getNullFunc() {
        return function() {};
    }
    private static getLogFunc(level:string, f, consoleRedirect) {
        return function(...args) {
            f.apply(consoleRedirect, [level].concat(args));
        };
    }
    redirectToFile() {
        this.wire(this.fileFuncs);
    }
    flushBufferToFile() {
        for(var i in this.logBuffer) {
            var entry = this.logBuffer[i];
            this.logToFile.apply(this, [entry.level, entry.message].concat(entry.stack));
        }
        this.logBuffer = [];
    }
    collect() {
        this.wire(this.collectFuns);
    }
    private static getStack() {
        var lines = new Error().stack.split("\n");
        
        if(lines.length > 4) {
            return lines.slice(4);
        }
        
        return lines;
    }
    flushBuffer() {
        var b = this.logBuffer;
        this.logBuffer = [];
        return b;
    }
}

function resolveComponent(name, components) {
	if(name.length == 0) {
		return undefined;
	}
	if(name.length == 1) {
		if(components.hasOwnProperty(name[0])) {
			return components[name[0]];			
		}
		return undefined;
	}
	if(components.hasOwnProperty(name[0])) {
		return resolveComponent(name.slice(1), components[name[0]]);						
	}
	return undefined;
}

function executeCall(components:any , callBuffer:Buffer, consoleRedirect:ConsoleRedirect):{
	    result:any;
	    error:string;
	    log:{
		    level:   string;
		    message: string;
		    stack:string[];
	    }[];   
    } {
        var call:{func:string;args:any[];};
        var error = "";
        var result = undefined;
        try {
            call = JSON.parse(callBuffer.toString());
            var componentFunc = resolveComponent(call.func.split("."), components);
            if(componentFunc === undefined) {
                error = "component func: \"" + call.func + "\" not found";
            } else {
                try {
                    consoleRedirect.collect();
                    result = componentFunc.apply(null, call.args);
                } catch(e) {
                    error = "could not call func:" + e.message;
                }
                consoleRedirect.redirectToFile();
            }
        } catch(e) {
            // json parsing failed
            error = "could not parse incloming json: " + e.message;
        }
        return {
            result : result,
            error  : error,
            log    : consoleRedirect.flushBuffer()
        }
}


function run(components, consoleRedirect:ConsoleRedirect) {
    var readBuffer = new Buffer("");
    process.stdin.on('readable', () => {
        var buffer:Buffer = process.stdin.read() as Buffer; // 38{"func": "foo.bar", "args": ["hallo"]}
        
        if(buffer) {
            readBuffer = Buffer.concat([readBuffer, buffer]);
        }

        // find start
        var callLength = 0;
        var callStart = 0;
        for(var i=0;i<readBuffer.length;i++) {
            if(readBuffer[i] == 123) { // opening "{"
                callStart = i;
                callLength = parseInt(readBuffer.slice(0, callStart).toString());
                break; 
            }
        }
        
        if(callLength > 0 && (readBuffer.length >= (callStart + callLength))) {
            // got a valid call
            var callResult = executeCall(components, readBuffer.slice(callStart, callStart+callLength), consoleRedirect);
            var resultBuffer = new Buffer(JSON.stringify(callResult));
            process.stdout.write(resultBuffer.length.toString());
            process.stdout.write(resultBuffer);
            
            readBuffer = readBuffer.slice(callStart+callLength);
        }
    });
    process.stdin.on('end', () => {
        process.stdout.write('std in was closed with a readbuffer ' + readBuffer.toString());
        process.exit(1)
    });
}

var ____consoleRedirect = new ConsoleRedirect(true, "/tmp/test.log", console);
____consoleRedirect.collect();
var jsSource = process.argv.pop();
var contents = fs.readFileSync(jsSource).toString();
var js = "(function(require, module) {" + contents + ";})(undefined, undefined);";
eval(js);
____consoleRedirect.flushBufferToFile();
____consoleRedirect.redirectToFile();
run(global, ____consoleRedirect);

