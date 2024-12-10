#!/bin/bash
#
#SBATCH --mail-user=XXX@XXX.com
#SBATCH --mail-type=ALL
#SBATCH --job-name=test_score
#SBATCH --output=/parallel-server/parserver/test/slurm/out/%j.%N.stdout
#SBATCH --error=/parallel-server/parserver/test/slurm/err/%j.%N.stderr
#SBATCH --chdir=/parallel-server/parserver/test
#SBATCH --partition=fast
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=48
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=10:00

module load golang/1.22
go run test.go parserver
