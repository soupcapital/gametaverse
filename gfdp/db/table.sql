CREATE TABLE t_txs (tx_hash String NOT NULL,
  ts DateTime,
  blk_num UInt64,
  from String,
  to String, 
  data String
)   ENGINE = ReplacingMergeTree()
    ORDER BY  (blk_num,ts, to, tx_hash)
    PRIMARY KEY (blk_num,ts,to);