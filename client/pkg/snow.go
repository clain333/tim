package pkg

import "github.com/bwmarrin/snowflake"

var SnowflakeNode *snowflake.Node

func Init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	SnowflakeNode = node
}
