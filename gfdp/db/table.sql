CREATE TABLE t_tx_bsc (
  ts DateTime,
  blk_num UInt64,
  tx_idx UInt16,
  from String,
  to String 
)   ENGINE = ReplacingMergeTree()
    ORDER BY  (blk_num, tx_idx, ts, to, from)
    PRIMARY KEY (blk_num, tx_idx, ts,to, from);