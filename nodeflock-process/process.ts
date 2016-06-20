
declare var process:any;
declare var require:any;

var fs = require('fs');
function testLog(entry) {
    fs.appendFile("test.log", "this is in the log file " + entry + "\n", function (err) { 
        if(err) {
            return console.error("could not write to file", err);
        }
    });
}

testLog("starting " + new Date());

//console.log("starting process", process.argv);

process.stdin.setEncoding('utf8');
function run() {
    process.stdin.on('readable', () => {
        var chunk = process.stdin.read();
        if (chunk !== null) {
            var debug = `data: ${chunk}\n`
            process.stdout.write(debug);
            testLog(debug);
        }
    });
    process.stdin.on('end', () => {
        process.stdout.write('std in was closed');
        run();
    });
}

run();

