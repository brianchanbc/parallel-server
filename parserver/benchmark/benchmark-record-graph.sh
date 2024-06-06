#!/bin/bash
#
#SBATCH --mail-user=XXX@XXX.com
#SBATCH --mail-type=ALL
#SBATCH --job-name=parserver_benchmark_record_graph
#SBATCH --output=/parallel-server/parserver/benchmark/slurm/out/speed_output.txt
#SBATCH --error=/parallel-server/parserver/benchmark/slurm/err/speed_output_err.txt
#SBATCH --chdir=/parallel-server/parserver/benchmark
#SBATCH --partition=fast
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=48
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=6:00:00

module load golang/1.22

SIZES=("xsmall" "small" "medium" "large" "xlarge")
THREADS=(2 4 6 8 12)

for size in "${SIZES[@]}"
do
    for i in {1..5}
    do
        output=$(go run benchmark.go s $size)
        echo "s,$size,$output"
    done
done

for size in "${SIZES[@]}"
do
    for threads in "${THREADS[@]}"
    do
        for i in {1..5}
        do
            output=$(go run benchmark.go p $size $threads)
            echo "p,$size,$threads,$output"
        done
    done
done

sleep 10

python graph.py