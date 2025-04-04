package converter

var UpdateFileName = `
	UPDATE video_table
	SET hls_file_name = $1,
		convert_status = $2
	WHERE video_seq = $3 AND
		user_id = $4
`

var UpdateConvertStatus = `
	UPDATE video_table
	SET convert_status = $1
	WHERE video_seq = $2 AND
		user_id = $3
`
