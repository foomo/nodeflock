#!/usr/bin/env node
/// <reference path="node.d.ts" />
var fs = require('fs');

declare var ReactDOMServer:any;

function log(...args) {
    return;
    /*
    for(var i in args) {
        fs.appendFile("/tmp/test.log", (new Date) + "" + JSON.stringify(args[i]) + "\n", function (err) { 
            if(err) {
                return console.error("could not write to file", err);
            }
        });
    }
    */
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

function executeCall(components:any , callBuffer:Buffer):{
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
                    result = componentFunc.apply(null, call.args);
                } catch(e) {
                    error = e.message;
                }
            }
        } catch(e) {
            // json parsing failed
            error = e.message;
        }
        return {
            result: result,
            error: error,
            log:[]
        }
}


function run(components) {
    var readBuffer = new Buffer("");
    process.stdin.on('readable', () => {
        var buffer:Buffer = process.stdin.read() as Buffer; // 38{"func": "foo.bar", "args": ["hallo"]}
        
        if(buffer) {
            log(buffer.toString());            
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
            var callResult = executeCall(components, readBuffer.slice(callStart, callStart+callLength));
            var resultBuffer = new Buffer(JSON.stringify(callResult));
            process.stdout.write(resultBuffer.length.toString());
            process.stdout.write(resultBuffer);
            log("replying", resultBuffer.toString());
            readBuffer = readBuffer.slice(callStart+callLength);
        }
    });
    process.stdin.on('end', () => {
        process.stdout.write('std in was closed with a readbuffer ' + readBuffer.toString());
        process.exit(1)
    });
}

var jsSource = process.argv.pop();
var contents = fs.readFileSync(jsSource).toString();
var js = "(function(require, module) {" + contents + ";})(undefined, undefined);";
eval(js);
run(global);

