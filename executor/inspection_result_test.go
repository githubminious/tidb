// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package executor_test

import (
	"context"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/domain"
	"github.com/pingcap/tidb/infoschema"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/session"
	"github.com/pingcap/tidb/sessionctx/variable"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/testkit"
)

var _ = SerialSuites(&inspectionResultSuite{})

type inspectionResultSuite struct {
	store kv.Storage
	dom   *domain.Domain
}

func (s *inspectionResultSuite) SetUpSuite(c *C) {
	store, dom, err := newStoreWithBootstrap()
	c.Assert(err, IsNil)
	s.store = store
	s.dom = dom
}

func (s *inspectionResultSuite) TearDownSuite(c *C) {
	s.dom.Close()
	s.store.Close()
}

func (s *inspectionResultSuite) TestInspectionResult(c *C) {
	tk := testkit.NewTestKitWithInit(c, s.store)

	mockData := map[string]variable.TableSnapshot{}
	// mock configuration inconsistent
	mockData[infoschema.TableClusterConfig] = variable.TableSnapshot{
		Rows: [][]types.Datum{
			types.MakeDatums("tidb", "192.168.3.22:4000", "ddl.lease", "1"),
			types.MakeDatums("tidb", "192.168.3.23:4000", "ddl.lease", "2"),
			types.MakeDatums("tidb", "192.168.3.24:4000", "ddl.lease", "1"),
			types.MakeDatums("tidb", "192.168.3.25:4000", "ddl.lease", "1"),
			types.MakeDatums("tidb", "192.168.3.24:4000", "log.slow-threshold", "0"),
			types.MakeDatums("tidb", "192.168.3.25:4000", "log.slow-threshold", "1"),
			types.MakeDatums("tikv", "192.168.3.32:26600", "coprocessor.high", "8"),
			types.MakeDatums("tikv", "192.168.3.33:26600", "coprocessor.high", "8"),
			types.MakeDatums("tikv", "192.168.3.34:26600", "coprocessor.high", "7"),
			types.MakeDatums("tikv", "192.168.3.35:26600", "coprocessor.high", "7"),
			types.MakeDatums("pd", "192.168.3.32:2379", "scheduler.limit", "3"),
			types.MakeDatums("pd", "192.168.3.33:2379", "scheduler.limit", "3"),
			types.MakeDatums("pd", "192.168.3.34:2379", "scheduler.limit", "3"),
			types.MakeDatums("pd", "192.168.3.35:2379", "scheduler.limit", "3"),
		},
	}
	// mock version inconsistent
	mockData[infoschema.TableClusterInfo] = variable.TableSnapshot{
		Rows: [][]types.Datum{
			types.MakeDatums("tidb", "192.168.1.11:1234", "192.168.1.11:1234", "4.0", "a234c"),
			types.MakeDatums("tidb", "192.168.1.12:1234", "192.168.1.11:1234", "4.0", "a234d"),
			types.MakeDatums("tidb", "192.168.1.13:1234", "192.168.1.11:1234", "4.0", "a234e"),
			types.MakeDatums("tikv", "192.168.1.21:1234", "192.168.1.21:1234", "4.0", "c234d"),
			types.MakeDatums("tikv", "192.168.1.22:1234", "192.168.1.22:1234", "4.0", "c234d"),
			types.MakeDatums("tikv", "192.168.1.23:1234", "192.168.1.23:1234", "4.0", "c234e"),
			types.MakeDatums("pd", "192.168.1.31:1234", "192.168.1.31:1234", "4.0", "m234c"),
			types.MakeDatums("pd", "192.168.1.32:1234", "192.168.1.32:1234", "4.0", "m234d"),
			types.MakeDatums("pd", "192.168.1.33:1234", "192.168.1.33:1234", "4.0", "m234e"),
		},
	}
	// mock load
	mockData[infoschema.TableClusterLoad] = variable.TableSnapshot{
		Rows: [][]types.Datum{
			types.MakeDatums("tidb", "192.168.1.11:1234", "memory", "virtual", "used-percent", "0.8"),
			types.MakeDatums("tidb", "192.168.1.12:1234", "memory", "virtual", "used-percent", "0.6"),
			types.MakeDatums("tidb", "192.168.1.13:1234", "memory", "swap", "used-percent", "0"),
			types.MakeDatums("tikv", "192.168.1.21:1234", "memory", "swap", "used-percent", "0.6"),
			types.MakeDatums("pd", "192.168.1.31:1234", "cpu", "cpu", "load1", "1.0"),
			types.MakeDatums("pd", "192.168.1.32:1234", "cpu", "cpu", "load5", "2.0"),
			types.MakeDatums("pd", "192.168.1.33:1234", "cpu", "cpu", "load15", "8.0"),
		},
	}
	mockData[infoschema.TableClusterHardware] = variable.TableSnapshot{
		Rows: [][]types.Datum{
			types.MakeDatums("tikv", "192.168.1.22:1234", "disk", "sda", "used-percent", "80"),
			types.MakeDatums("tikv", "192.168.1.23:1234", "disk", "sdb", "used-percent", "50"),
			types.MakeDatums("pd", "192.168.1.31:1234", "cpu", "cpu", "cpu-logical-cores", "1"),
			types.MakeDatums("pd", "192.168.1.32:1234", "cpu", "cpu", "cpu-logical-cores", "4"),
			types.MakeDatums("pd", "192.168.1.33:1234", "cpu", "cpu", "cpu-logical-cores", "10"),
		},
	}

	ctx := context.WithValue(context.Background(), "__mockInspectionTables", mockData)
	fpName := "github.com/pingcap/tidb/executor/mockMergeMockInspectionTables"
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname string) bool {
		return fpname == fpName
	})

	c.Assert(failpoint.Enable(fpName, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName), IsNil) }()

	cases := []struct {
		sql  string
		rows []string
	}{
		{
			sql: "select rule, item, type, value, reference, severity, details from information_schema.inspection_result where rule in ('config', 'version')",
			rows: []string{
				"config coprocessor.high tikv inconsistent consistent warning the cluster has different config value of coprocessor.high, execute the sql to see more detail: select * from information_schema.cluster_config where type='tikv' and `key`='coprocessor.high'",
				"config ddl.lease tidb inconsistent consistent warning the cluster has different config value of ddl.lease, execute the sql to see more detail: select * from information_schema.cluster_config where type='tidb' and `key`='ddl.lease'",
				"config log.slow-threshold tidb 0 not 0 warning slow-threshold = 0 will record every query to slow log, it may affect performance",
				"config log.slow-threshold tidb inconsistent consistent warning the cluster has different config value of log.slow-threshold, execute the sql to see more detail: select * from information_schema.cluster_config where type='tidb' and `key`='log.slow-threshold'",
				"version git_hash tidb inconsistent consistent critical the cluster has 3 different tidb versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='tidb'",
				"version git_hash tikv inconsistent consistent critical the cluster has 2 different tikv versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='tikv'",
				"version git_hash pd inconsistent consistent critical the cluster has 3 different pd versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='pd'",
			},
		},
		{
			sql: "select rule, item, type, value, reference, severity, details from information_schema.inspection_result where rule in ('config', 'version') and item in ('coprocessor.high', 'git_hash') and type='tikv'",
			rows: []string{
				"config coprocessor.high tikv inconsistent consistent warning the cluster has different config value of coprocessor.high, execute the sql to see more detail: select * from information_schema.cluster_config where type='tikv' and `key`='coprocessor.high'",
				"version git_hash tikv inconsistent consistent critical the cluster has 2 different tikv versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='tikv'",
			},
		},
		{
			sql: "select rule, item, type, value, reference, severity, details from information_schema.inspection_result where rule='config'",
			rows: []string{
				"config coprocessor.high tikv inconsistent consistent warning the cluster has different config value of coprocessor.high, execute the sql to see more detail: select * from information_schema.cluster_config where type='tikv' and `key`='coprocessor.high'",
				"config ddl.lease tidb inconsistent consistent warning the cluster has different config value of ddl.lease, execute the sql to see more detail: select * from information_schema.cluster_config where type='tidb' and `key`='ddl.lease'",
				"config log.slow-threshold tidb 0 not 0 warning slow-threshold = 0 will record every query to slow log, it may affect performance",
				"config log.slow-threshold tidb inconsistent consistent warning the cluster has different config value of log.slow-threshold, execute the sql to see more detail: select * from information_schema.cluster_config where type='tidb' and `key`='log.slow-threshold'",
			},
		},
		{
			sql: "select rule, item, type, value, reference, severity, details from information_schema.inspection_result where rule='version' and item='git_hash' and type in ('pd', 'tidb')",
			rows: []string{
				"version git_hash tidb inconsistent consistent critical the cluster has 3 different tidb versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='tidb'",
				"version git_hash pd inconsistent consistent critical the cluster has 3 different pd versions, execute the sql to see more detail: select * from information_schema.cluster_info where type='pd'",
			},
		},
		{
			sql: "select rule, item, type, instance, value, reference, severity, details from information_schema.inspection_result where rule='current-load'",
			rows: []string{
				"current-load cpu-load1 pd 192.168.1.31:1234 1.0 < 0.7 warning cpu-load1 should less than (cpu_logical_cores * 0.7)",
				"current-load cpu-load15 pd 192.168.1.33:1234 8.0 < 7.0 warning cpu-load15 should less than (cpu_logical_cores * 0.7)",
				"current-load disk-usage tikv 192.168.1.22:1234 80 < 70 warning current disk-usage is too high, execute the sql to see more detail: select * from information_schema.cluster_hardware where type='tikv' and instance='192.168.1.22:1234' and device_type='disk' and device_name='sda'",
				"current-load swap-memory-usage tikv 192.168.1.21:1234 0.6 0 warning ",
				"current-load virtual-memory-usage tidb 192.168.1.11:1234 0.8 < 0.7 warning ",
			},
		},
	}

	for _, cs := range cases {
		rs, err := tk.Se.Execute(ctx, cs.sql)
		c.Assert(err, IsNil)
		result := tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("SQL: %v", cs.sql))
		warnings := tk.Se.GetSessionVars().StmtCtx.GetWarnings()
		c.Assert(len(warnings), Equals, 0, Commentf("expected no warning, got: %+v", warnings))
		result.Check(testkit.Rows(cs.rows...))
	}
}

func (s *inspectionResultSuite) parseTime(c *C, se session.Session, str string) types.Time {
	t, err := types.ParseTime(se.GetSessionVars().StmtCtx, str, mysql.TypeDatetime, types.MaxFsp)
	c.Assert(err, IsNil)
	return t
}

func (s *inspectionResultSuite) TestThresholdCheckInspection(c *C) {
	tk := testkit.NewTestKitWithInit(c, s.store)
	// mock tikv configuration.
	configurations := map[string]variable.TableSnapshot{}
	configurations[infoschema.TableClusterConfig] = variable.TableSnapshot{
		Rows: [][]types.Datum{
			types.MakeDatums("tikv", "tikv-0", "raftstore.apply-pool-size", "2"),
			types.MakeDatums("tikv", "tikv-0", "raftstore.store-pool-size", "2"),
			types.MakeDatums("tikv", "tikv-0", "readpool.coprocessor.high-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "readpool.coprocessor.low-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "readpool.coprocessor.normal-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "readpool.storage.high-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "readpool.storage.low-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "readpool.storage.normal-concurrency", "4"),
			types.MakeDatums("tikv", "tikv-0", "server.grpc-concurrency", "8"),
			types.MakeDatums("tikv", "tikv-0", "storage.scheduler-worker-pool-size", "6"),
		},
	}
	datetime := func(str string) types.Time {
		return s.parseTime(c, tk.Se, str)
	}
	// construct some mock abnormal data
	mockData := map[string][][]types.Datum{
		// columns: time, instance, name, value
		"tikv_thread_cpu": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_normal0", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_normal1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_high1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_low1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "grpc_1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "raftstore_1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "apply_0", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_norm1", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_high2", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_low0", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "sched_2", 10.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "split_check", 10.0),
		},
		"pd_tso_wait_duration":                {},
		"tidb_get_token_duration":             {},
		"tidb_load_schema_duration":           {},
		"tikv_scheduler_command_duration":     {},
		"tikv_handle_snapshot_duration":       {},
		"tikv_storage_async_request_duration": {},
		"tikv_engine_write_duration":          {},
		"tikv_engine_max_get_duration":        {},
		"tikv_engine_max_seek_duration":       {},
		"tikv_scheduler_pending_commands":     {},
		"tikv_block_index_cache_hit":          {},
		"tikv_block_data_cache_hit":           {},
		"tikv_block_filter_cache_hit":         {},
		"pd_scheduler_store_status":           {},
		"pd_region_health":                    {},
	}

	fpName := "github.com/pingcap/tidb/executor/mockMergeMockInspectionTables"
	c.Assert(failpoint.Enable(fpName, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName), IsNil) }()

	// Mock for metric table data.
	fpName2 := "github.com/pingcap/tidb/executor/mockMetricsTableData"
	c.Assert(failpoint.Enable(fpName2, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName2), IsNil) }()

	ctx := context.WithValue(context.Background(), "__mockInspectionTables", configurations)
	ctx = context.WithValue(ctx, "__mockMetricsTableData", mockData)
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname2 string) bool {
		return fpName2 == fpname2 || fpname2 == fpName
	})

	rs, err := tk.Se.Execute(ctx, "select  /*+ time_range('2020-02-12 10:35:00','2020-02-12 10:37:00') */ item, type, instance, value, reference, details from information_schema.inspection_result where rule='threshold-check' order by item")
	c.Assert(err, IsNil)
	result := tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("execute inspect SQL failed"))
	c.Assert(tk.Se.GetSessionVars().StmtCtx.WarningCount(), Equals, uint16(0), Commentf("unexpected warnings: %+v", tk.Se.GetSessionVars().StmtCtx.GetWarnings()))
	result.Check(testkit.Rows(
		"apply-cpu tikv tikv-0 10.00 < 1.60, config: raftstore.apply-pool-size=2 the 'apply-cpu' max cpu-usage of tikv-0 tikv is too high",
		"coprocessor-high-cpu tikv tikv-0 10.00 < 3.60, config: readpool.coprocessor.high-concurrency=4 the 'coprocessor-high-cpu' max cpu-usage of tikv-0 tikv is too high",
		"coprocessor-low-cpu tikv tikv-0 10.00 < 3.60, config: readpool.coprocessor.low-concurrency=4 the 'coprocessor-low-cpu' max cpu-usage of tikv-0 tikv is too high",
		"coprocessor-normal-cpu tikv tikv-0 10.00 < 3.60, config: readpool.coprocessor.normal-concurrency=4 the 'coprocessor-normal-cpu' max cpu-usage of tikv-0 tikv is too high",
		"grpc-cpu tikv tikv-0 10.00 < 7.20, config: server.grpc-concurrency=8 the 'grpc-cpu' max cpu-usage of tikv-0 tikv is too high",
		"raftstore-cpu tikv tikv-0 10.00 < 1.60, config: raftstore.store-pool-size=2 the 'raftstore-cpu' max cpu-usage of tikv-0 tikv is too high",
		"scheduler-worker-cpu tikv tikv-0 10.00 < 5.10, config: storage.scheduler-worker-pool-size=6 the 'scheduler-worker-cpu' max cpu-usage of tikv-0 tikv is too high",
		"split-check-cpu tikv tikv-0 10.00 < 0.00 the 'split-check-cpu' max cpu-usage of tikv-0 tikv is too high",
		"storage-readpool-high-cpu tikv tikv-0 10.00 < 3.60, config: readpool.storage.high-concurrency=4 the 'storage-readpool-high-cpu' max cpu-usage of tikv-0 tikv is too high",
		"storage-readpool-low-cpu tikv tikv-0 10.00 < 3.60, config: readpool.storage.low-concurrency=4 the 'storage-readpool-low-cpu' max cpu-usage of tikv-0 tikv is too high",
		"storage-readpool-normal-cpu tikv tikv-0 10.00 < 3.60, config: readpool.storage.normal-concurrency=4 the 'storage-readpool-normal-cpu' max cpu-usage of tikv-0 tikv is too high",
	))

	// construct some mock normal data
	mockData["tikv_thread_cpu"] = [][]types.Datum{
		// columns: time, instance, name, value
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_normal0", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_high1", 0.1),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "cop_low1", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "grpc_1", 7.21),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "grpc_2", 0.21),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "raftstore_1", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "apply_0", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_norm1", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_high2", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "store_read_low0", 1.0),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "sched_2", 0.3),
		types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "split_check", 0.5),
	}

	ctx = context.WithValue(context.Background(), "__mockInspectionTables", configurations)
	ctx = context.WithValue(ctx, "__mockMetricsTableData", mockData)
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname2 string) bool {
		return fpName2 == fpname2 || fpname2 == fpName
	})
	rs, err = tk.Se.Execute(ctx, "select item, type, instance, value, reference from information_schema.inspection_result where rule='threshold-check' order by item")
	c.Assert(err, IsNil)
	result = tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("execute inspect SQL failed"))
	c.Assert(tk.Se.GetSessionVars().StmtCtx.WarningCount(), Equals, uint16(0), Commentf("unexpected warnings: %+v", tk.Se.GetSessionVars().StmtCtx.GetWarnings()))
	result.Check(testkit.Rows("grpc-cpu tikv tikv-0 7.21 < 7.20, config: server.grpc-concurrency=8"))
}

func (s *inspectionResultSuite) TestThresholdCheckInspection2(c *C) {
	tk := testkit.NewTestKitWithInit(c, s.store)
	// Mock for metric table data.
	fpName := "github.com/pingcap/tidb/executor/mockMetricsTableData"
	c.Assert(failpoint.Enable(fpName, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName), IsNil) }()

	datetime := func(s string) types.Time {
		t, err := types.ParseTime(tk.Se.GetSessionVars().StmtCtx, s, mysql.TypeDatetime, types.MaxFsp)
		c.Assert(err, IsNil)
		return t
	}

	// construct some mock abnormal data
	mockData := map[string][][]types.Datum{
		"pd_tso_wait_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", 0.999, 0.06),
		},
		"tidb_get_token_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tidb-0", 0.999, 0.02*10e5),
		},
		"tidb_load_schema_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tidb-0", 0.99, 2.0),
		},
		"tikv_scheduler_command_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "get", 0.99, 2.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "write", 0.99, 5.0),
		},
		"tikv_handle_snapshot_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "gen", 0.999, 40.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "read", 0.999, 10.0),
		},
		"tikv_storage_async_request_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "write", 0.999, 0.2),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "snapshot", 0.999, 0.06),
		},
		"tikv_engine_write_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "write_max", "kv", 0.2*10e5),
		},
		"tikv_engine_max_get_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "get_max", "kv", 0.06*10e5),
		},
		"tikv_engine_max_seek_duration": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "seek_max", "raft", 0.06*10e5),
		},
		"tikv_scheduler_pending_commands": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", 1001.0),
		},
		"tikv_block_index_cache_hit": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "kv", 0.94),
		},
		"tikv_block_data_cache_hit": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "kv", 0.79),
		},
		"tikv_block_filter_cache_hit": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "tikv-0", "kv", 0.93),
		},
		"tikv_thread_cpu":           {},
		"pd_scheduler_store_status": {},
		"pd_region_health":          {},
	}

	ctx := context.WithValue(context.Background(), "__mockMetricsTableData", mockData)
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname string) bool {
		return fpname == fpName
	})

	rs, err := tk.Se.Execute(ctx, "select /*+ time_range('2020-02-12 10:35:00','2020-02-12 10:37:00') */ item, type, instance, value, reference, details from information_schema.inspection_result where rule='threshold-check' order by item")
	c.Assert(err, IsNil)
	result := tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("execute inspect SQL failed"))
	c.Assert(tk.Se.GetSessionVars().StmtCtx.WarningCount(), Equals, uint16(0), Commentf("unexpected warnings: %+v", tk.Se.GetSessionVars().StmtCtx.GetWarnings()))
	result.Check(testkit.Rows(
		"data-block-cache-hit tikv tikv-0 0.790 > 0.800 min data-block-cache-hit rate of tikv-0 tikv is too low",
		"filter-block-cache-hit tikv tikv-0 0.930 > 0.950 min filter-block-cache-hit rate of tikv-0 tikv is too low",
		"get-token-duration tidb tidb-0 0.020 < 0.001 max duration of tidb-0 tidb get-token-duration is too slow",
		"handle-snapshot-duration tikv tikv-0 40.000 < 30.000 max duration of tikv-0 tikv handle-snapshot-duration is too slow",
		"index-block-cache-hit tikv tikv-0 0.940 > 0.950 min index-block-cache-hit rate of tikv-0 tikv is too low",
		"load-schema-duration tidb tidb-0 2.000 < 1.000 max duration of tidb-0 tidb load-schema-duration is too slow",
		"rocksdb-get-duration tikv tikv-0 0.060 < 0.050 max duration of tikv-0 tikv rocksdb-get-duration is too slow",
		"rocksdb-seek-duration tikv tikv-0 0.060 < 0.050 max duration of tikv-0 tikv rocksdb-seek-duration is too slow",
		"rocksdb-write-duration tikv tikv-0 0.200 < 0.100 max duration of tikv-0 tikv rocksdb-write-duration is too slow",
		"scheduler-cmd-duration tikv tikv-0 5.000 < 0.100 max duration of tikv-0 tikv scheduler-cmd-duration is too slow",
		"scheduler-pending-cmd-count tikv tikv-0 1001.000 < 1000.000  tikv-0 tikv scheduler has too many pending commands",
		"storage-snapshot-duration tikv tikv-0 0.060 < 0.050 max duration of tikv-0 tikv storage-snapshot-duration is too slow",
		"storage-write-duration tikv tikv-0 0.200 < 0.100 max duration of tikv-0 tikv storage-write-duration is too slow",
		"tso-duration tidb pd-0 0.060 < 0.050 max duration of pd-0 tidb tso-duration is too slow",
	))
}

func (s *inspectionResultSuite) TestThresholdCheckInspection3(c *C) {
	tk := testkit.NewTestKitWithInit(c, s.store)
	// Mock for metric table data.
	fpName := "github.com/pingcap/tidb/executor/mockMetricsTableData"
	c.Assert(failpoint.Enable(fpName, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName), IsNil) }()

	datetime := func(s string) types.Time {
		t, err := types.ParseTime(tk.Se.GetSessionVars().StmtCtx, s, mysql.TypeDatetime, types.MaxFsp)
		c.Assert(err, IsNil)
		return t
	}

	// construct some mock abnormal data
	mockData := map[string][][]types.Datum{
		"pd_scheduler_store_status": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-0", "0", "leader_score", 100.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-1", "1", "leader_score", 50.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-0", "0", "region_score", 100.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-1", "1", "region_score", 90.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-0", "0", "store_available", 100.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-1", "1", "store_available", 70.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "tikv-0", "0", "region_count", 20001.0),
		},
		"pd_region_health": {
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "extra-peer-region-count", 40.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "learner-peer-region-count", 40.0),
			types.MakeDatums(datetime("2020-02-14 05:20:00"), "pd-0", "pending-peer-region-count", 30.0),
		},
	}

	ctx := context.WithValue(context.Background(), "__mockMetricsTableData", mockData)
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname string) bool {
		return fpname == fpName
	})

	rs, err := tk.Se.Execute(ctx, `select /*+ time_range('2020-02-14 04:20:00','2020-02-14 05:20:00') */
		item, type, instance, value, reference, details from information_schema.inspection_result
		where rule='threshold-check' and item in ('leader-score-balance','region-score-balance','region-count','region-health','store-available-balance')
		order by item`)
	c.Assert(err, IsNil)
	result := tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("execute inspect SQL failed"))
	c.Assert(tk.Se.GetSessionVars().StmtCtx.WarningCount(), Equals, uint16(0), Commentf("unexpected warnings: %+v", tk.Se.GetSessionVars().StmtCtx.GetWarnings()))
	result.Check(testkit.Rows(
		"leader-score-balance tikv tikv-1 50.00% < 5.00% tikv-0 leader_score is 100.00, much more than tikv-1 leader_score 50.00",
		"region-count tikv tikv-0 20001.00 <= 20000 tikv-0 tikv has too many regions",
		"region-health pd pd-0 110.00 < 100 the count of extra-perr and learner-peer and pending-peer are 110, it means the scheduling is too frequent or too slow",
		"region-score-balance tikv tikv-1 10.00% < 5.00% tikv-0 region_score is 100.00, much more than tikv-1 region_score 90.00",
		"store-available-balance tikv tikv-1 30.00% < 20.00% tikv-0 store_available is 100.00, much more than tikv-1 store_available 70.00"))
}

func (s *inspectionResultSuite) TestCriticalErrorInspection(c *C) {
	tk := testkit.NewTestKitWithInit(c, s.store)

	fpName := "github.com/pingcap/tidb/executor/mockMetricsTableData"
	c.Assert(failpoint.Enable(fpName, "return"), IsNil)
	defer func() { c.Assert(failpoint.Disable(fpName), IsNil) }()

	datetime := func(str string) types.Time {
		return s.parseTime(c, tk.Se, str)
	}

	// construct some mock data
	mockData := map[string][][]types.Datum{
		// columns: time, instance, type, value
		"tikv_critical_error_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tikv-0", "type1", 0.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tikv-1", "type1", 1.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tikv-2", "type2", 5.0),
		},
		// columns: time, instance, value
		"tidb_panic_count_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tidb-0", 4.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tidb-0", 0.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tidb-1", 1.0),
		},
		// columns: time, instance, value
		"tidb_binlog_error_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tidb-1", 4.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tidb-2", 0.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tidb-3", 1.0),
		},
		// columns: time, instance, db, type, stage, value
		"tikv_scheduler_is_busy_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tikv-0", "db1", "type1", "stage1", 1.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tikv-0", "db2", "type1", "stage2", 2.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tikv-1", "db1", "type2", "stage1", 3.0),
			types.MakeDatums(datetime("2020-02-12 10:38:00"), "tikv-0", "db1", "type1", "stage2", 4.0),
			types.MakeDatums(datetime("2020-02-12 10:39:00"), "tikv-0", "db2", "type1", "stage1", 5.0),
			types.MakeDatums(datetime("2020-02-12 10:40:00"), "tikv-1", "db1", "type2", "stage2", 6.0),
		},
		// columns: time, instance, db, value
		"tikv_coprocessor_is_busy_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tikv-0", "db1", 1.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tikv-0", "db2", 2.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tikv-1", "db1", 3.0),
			types.MakeDatums(datetime("2020-02-12 10:38:00"), "tikv-0", "db1", 4.0),
			types.MakeDatums(datetime("2020-02-12 10:39:00"), "tikv-0", "db2", 5.0),
			types.MakeDatums(datetime("2020-02-12 10:40:00"), "tikv-1", "db1", 6.0),
		},
		// columns: time, instance, db, type, value
		"tikv_channel_full_total_count": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tikv-0", "db1", "type1", 1.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tikv-0", "db2", "type1", 2.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tikv-1", "db1", "type2", 3.0),
			types.MakeDatums(datetime("2020-02-12 10:38:00"), "tikv-0", "db1", "type1", 4.0),
			types.MakeDatums(datetime("2020-02-12 10:39:00"), "tikv-0", "db2", "type1", 5.0),
			types.MakeDatums(datetime("2020-02-12 10:40:00"), "tikv-1", "db1", "type2", 6.0),
		},
		// columns: time, instance, db, value
		"tikv_engine_write_stall": {
			types.MakeDatums(datetime("2020-02-12 10:35:00"), "tikv-0", "kv", 1.0),
			types.MakeDatums(datetime("2020-02-12 10:36:00"), "tikv-0", "raft", 2.0),
			types.MakeDatums(datetime("2020-02-12 10:37:00"), "tikv-1", "reason3", 3.0),
		},
	}

	ctx := context.WithValue(context.Background(), "__mockMetricsTableData", mockData)
	ctx = failpoint.WithHook(ctx, func(_ context.Context, fpname string) bool {
		return fpName == fpname
	})

	rs, err := tk.Se.Execute(ctx, "select /*+ time_range('2020-02-12 10:35:00','2020-02-12 10:37:00') */ item, instance, value, details from information_schema.inspection_result where rule='critical-error'")
	c.Assert(err, IsNil)
	result := tk.ResultSetToResultWithCtx(ctx, rs[0], Commentf("execute inspect SQL failed"))
	c.Assert(tk.Se.GetSessionVars().StmtCtx.WarningCount(), Equals, uint16(0), Commentf("unexpected warnings: %+v", tk.Se.GetSessionVars().StmtCtx.GetWarnings()))
	result.Check(testkit.Rows(
		"channel-is-full tikv-1 9.00(db1, type2) the total number of errors about 'channel-is-full' is too many",
		"coprocessor-is-busy tikv-1 9.00(db1) the total number of errors about 'coprocessor-is-busy' is too many",
		"channel-is-full tikv-0 7.00(db2, type1) the total number of errors about 'channel-is-full' is too many",
		"coprocessor-is-busy tikv-0 7.00(db2) the total number of errors about 'coprocessor-is-busy' is too many",
		"scheduler-is-busy tikv-1 6.00(db1, type2, stage2) the total number of errors about 'scheduler-is-busy' is too many",
		"channel-is-full tikv-0 5.00(db1, type1) the total number of errors about 'channel-is-full' is too many",
		"coprocessor-is-busy tikv-0 5.00(db1) the total number of errors about 'coprocessor-is-busy' is too many",
		"critical-error tikv-2 5.00(type2) the total number of errors about 'critical-error' is too many",
		"scheduler-is-busy tikv-0 5.00(db2, type1, stage1) the total number of errors about 'scheduler-is-busy' is too many",
		"binlog-error tidb-1 4.00 the total number of errors about 'binlog-error' is too many",
		"panic-count tidb-0 4.00 the total number of errors about 'panic-count' is too many",
		"scheduler-is-busy tikv-0 4.00(db1, type1, stage2) the total number of errors about 'scheduler-is-busy' is too many",
		"scheduler-is-busy tikv-1 3.00(db1, type2, stage1) the total number of errors about 'scheduler-is-busy' is too many",
		"tikv_engine_write_stall tikv-1 3.00(reason3) the total number of errors about 'tikv_engine_write_stall' is too many",
		"scheduler-is-busy tikv-0 2.00(db2, type1, stage2) the total number of errors about 'scheduler-is-busy' is too many",
		"tikv_engine_write_stall tikv-0 2.00(raft) the total number of errors about 'tikv_engine_write_stall' is too many",
		"binlog-error tidb-3 1.00 the total number of errors about 'binlog-error' is too many",
		"critical-error tikv-1 1.00(type1) the total number of errors about 'critical-error' is too many",
		"panic-count tidb-1 1.00 the total number of errors about 'panic-count' is too many",
		"scheduler-is-busy tikv-0 1.00(db1, type1, stage1) the total number of errors about 'scheduler-is-busy' is too many",
		"tikv_engine_write_stall tikv-0 1.00(kv) the total number of errors about 'tikv_engine_write_stall' is too many",
	))
}
