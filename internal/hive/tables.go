package hive

type Table struct {
	Database, Table string
	Schemas, DDL    string
	PartitionCols   []string
	Partitions      []string
}
