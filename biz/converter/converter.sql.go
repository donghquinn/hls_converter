package converter

var InsertConvertedFileName = `
	UPDATE video_table
	SET hls_file_name = $1,
		convert_status = $2
	WHERE video_seq = $3
`
