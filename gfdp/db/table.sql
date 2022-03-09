CREATE TABLE t_tx_bsc (
  ts DateTime,
  blk_num UInt64,
  from String,
  to String 
)   ENGINE = ReplacingMergeTree()
    ORDER BY  (blk_num,ts, to)
    PRIMARY KEY (blk_num,ts,to);