const { execSync } = require("child_process");
const { writeFileSync } = require("fs");

const NS_PER_SEC = 1e9;

const P = ["BSP_off", "BSP_on"];
const N = [1, 2, 4, 6, 8];

const run_twitter_par = (bsp, timing, threads) => {
    const time = process.hrtime();
    try {
	    execSync(`./src/driver/driver ${threads ? `-p=${threads}` : ""} ${bsp == "BSP_on" ? "-bsp" : ""} < src/test.txt > /dev/null`); // executable must have been built with go build
    } catch (err) {
        console.log(`Timing #${timing} FAILED`);
        return -1;
    }
    const [secs, ns] = process.hrtime(time);
    console.log(`Timing #${timing}: ${secs}s + ${ns}ns`);
    return secs * NS_PER_SEC + ns;
};

const run_twitter_seq = (bsp, timing) => run_twitter_par(bsp, timing);

const run_test = () => {
    const seq_results = {};
    const par_results = {};

    // RUN SEQUENTIAL
    let timing = 1;
    const timetaken = run_twitter_seq("BSP_off", timing);
    seq_results["BSP_off"] = timetaken;
    timing++;

    // RUN PARALLEL
    P.forEach(bsp => {
        N.forEach(threads => {
            const timetaken = run_twitter_par(bsp, timing, threads);
            par_results[bsp] = {
                [threads]: timetaken,
                ...par_results[bsp]
            };
            timing++;
        });
    });

    console.log(seq_results);
    console.log(par_results);
    return [seq_results, par_results]
};

const compute_avg_par = tests => {
    const sums = tests.reduce((acc, cur) => {
        for (const key in cur) {
            if (key in acc)
                for (const threads in cur[key])
                    acc[key][threads] += cur[key][threads];
            else
                acc[key] = cur[key];
        }
        return acc;
    }, {});

    for (const key in sums)
        for (const threads in sums[key])
            sums[key][threads] /= tests.length;
    return sums;
}

const compute_avg_seq = tests => {
    const sums = tests.reduce((acc, cur) => {
        for (const key in cur)
            key in acc ? acc[key] += cur[key] : acc[key] = cur[key];
        return acc;
    }, {});

    for (const key in sums)
        sums[key] /= tests.length;
    return sums;
}

const run = () => {
    const seq_tests = [];
    const par_tests = [];
    for (let i=0; i<20; i++) {
        const [seq, par] = run_test();
        seq_tests.push(seq);
        par_tests.push(par);
    }

    const seq_avg = compute_avg_seq(seq_tests);
    const par_avg = compute_avg_par(par_tests);

    writeFileSync("timing_results_sequential.json", JSON.stringify(seq_avg));
    writeFileSync("timing_results_parallel.json", JSON.stringify(par_avg));
};

run();
