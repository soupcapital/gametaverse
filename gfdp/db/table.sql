CREATE TABLE t_txs (
  ts DateTime,
  blk_num UInt64,
  from String,
  to String 
)   ENGINE = ReplacingMergeTree()
    ORDER BY  (blk_num,ts, to)
    PRIMARY KEY (blk_num,ts,to);