package configkafka
// var w := kafka.NewWriter(kafka.WriterConfig{
// 		Brokers: []string{"localhost:9092"},
// 		Topic: "task.events",
// 		RequiredAcks: int(kafka.RequireAll), // гарантия доставки для надежности
// 		BatchSize: batchSize,
// 		BatchTimeout: 10*time.Millisecond,
// 		MaxAttempts: 3,
// 		Async: false, 
// })