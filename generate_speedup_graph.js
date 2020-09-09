// Code modified from source: https://sweetcode.io/nodejs-highcharts-sweetcode/

const fs = require("fs");
const chartExporter = require("highcharts-export-server");

const BSP_OFF = "BSP_off";
const BSP_ON = "BSP_on";

const coords = timings =>
    Object.entries(timings).map(timing => {
        const [threads, speedup] = timing;
        return {
            x: parseInt(threads),
            y: speedup
        };
    });

// Chart details object specifies chart type and data to plot
const generateDetails = details => ({
   type: "png",
   options: {
        title: {
            text: "Speedup vs Worker Threads"
        },
        legend: {
            title: { text: "File (P)" }
        },
        series: [
            {
                name: BSP_OFF,
                type: "line",
                data: coords(details[BSP_OFF])
            },
            {
                name: BSP_ON,
                type: "line",
                data: coords(details[BSP_ON])
            },
        ],
        xAxis: {
            max: 9,
            title: { text: "Number of worker threads (N)" }
        },
        yAxis: {
            title: { text: "Speedup amount" }
        }
    }
});

const exportGraph = chartDetails => {
    // Initialize the exporter
    chartExporter.initPool();

    chartExporter.export(chartDetails, (err, res) => {
        // Get the image data (base64)
        let imageb64 = res.data;
        
        // Filename of the output
        let outputFile = "speedup_graph.png";
        
        // Save the image to file
        fs.writeFileSync(outputFile, imageb64, "base64", function(err) {
            if (err) console.log(err);
        });
    
        console.log("Saved image!");
        chartExporter.killPool();
    })
};

const deepCopy = obj => JSON.parse(JSON.stringify(obj));

const getResults = () => {
    const seq = JSON.parse(fs.readFileSync("timing_results_sequential.json"));
    const par = JSON.parse(fs.readFileSync("timing_results_parallel.json"));
    return [seq, par];
};

const calcSpeedup = (seq, par) => {
    // Speedup = serial execution / parallel execution
    const speedup = deepCopy(par);
    for (const commands in speedup)
        for (const threads in speedup[commands])
            speedup[commands][threads] = seq["BSP_off"] / par[commands][threads];
    return speedup;

};

const generateGraph = () => {
    const [seq, par] = getResults();
    const speedup = calcSpeedup(seq, par);
    const chartDetails = generateDetails(speedup);
    exportGraph(chartDetails);
};

generateGraph();