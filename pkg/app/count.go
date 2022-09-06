package app

func (r *App) CountMedia() (int, error) {
	nextToken := "1"
	count := 0
	for nextToken != "" {
		if nextToken == "1" {
			nextToken = ""
		}
		token, items, err := r.googlePhotoClient.ListMediaItems(100, nextToken)
		if err != nil {
			return 0, err
		}
		nextToken = token
		count += len(items)
	}
	return count, nil
}
