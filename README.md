# fastq_demultiplexer

Converts fastq single cell data from MGI (10x converted library) to Illumina compatible format.

---

## Installation

```shell
go install github.com/overerd/fastq_demultiplexer@latest
```

---

### Arguments

#### Required:
* `-1|--r1` and `-2|--r2` expect an R1-R2 pair of fastq files.
* `-c|--csv-file` table with barcodes for each index.
* `-o|--output-directory` output directory (would contain subdirectories of demultiplexed samples).

#### Optional:
* `-s|--csv-separator` barcode csv-file separator (default: ',').
* `--targets-file` path to file with targets (if null, would select all possible indexes from barcodes file).
* `--transform-strategy` strategy of how to transform fastq data (supported strategies: 10x, 10x_no_index) (default: 10x)
* `--filename-template` filename template (default: '{{.SampleName}}_S{{.SampleNumber}}_L00{{.LaneNumber}}_{{.ReadType}}_001.fastq.gz').
* `--lane-number` lane number for selected fastq pair (default: 1).
* `--buffer-size` sets buffer size in bytes for reading fastq files (should be set and increased if necessary to avoid "take too long" error when reading fastq files with long lines) (default: 10 * 1024 * 1024).
* `--block-size` sets buffer size for accumulating multiple paired reads in transformation pipeline (default: 4 * 2 * 1024).
* `--compression-level` output gzip compression level if applicable [1, 9] (default: 1).
* `--debug` enables debug messages.

---

#### Filename template:

It uses golang template syntax `{{.Variable}}`.

Template `{{.SampleName}}_S{{.SampleNumber}}_L00{{.LaneNumber}}_{{.ReadType}}_001.fastq.gz` would result in filenames like `H2_S1_L001_R2_001.fastq.gz`.

##### Supported template variables:

* `{{.SampleName}}` - index name from barcodes csv-file
* `{{.SampleNumber}}` - index number
* `{{.LaneNumber}}` - `--lane-number value`
* `{{.ReadType}}` - read type (could be R1, R2 and I1)

---

### Example

```shell
fastq_demultiplexer \
    -1 v350013347_run65_L01_read_1.fq.gz \
    -2 v350013347_run65_L01_read_2.fq.gz \
    -c Single_Index_Kit_T_Set_A.csv \
    --lane-number 1 \
    --block-size 1000 \
    --targets-file targets.txt \
    --transform-strategy 10x_no_index \
    --filename-template {{.SampleName}}_S{{.SampleNumber}}_L00{{.LaneNumber}}_{{.ReadType}}_001.fastq.gz \
    -o output/ \
    --debug
```
