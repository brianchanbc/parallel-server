#!/bin/bash
#
#SBATCH --mail-user=XXX@XXX.com
#SBATCH --mail-type=ALL
#SBATCH --job-name=parserver_benchmark 
#SBATCH --output=/parallel-server/parserver/test/slurm/out/%j.%N.stdout
#SBATCH --error=/parallel-server/parserver/test/slurm/err/%j.%N.stderr
#SBATCH --chdir=/parallel-server/parserver
#SBATCH --partition=fast 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=48
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=5:00


module load golang/1.22
go test parserver/twitter -v -count=1
