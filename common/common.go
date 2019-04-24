package common

/*
func addAlbumToAlbumList(userID, albumID string) int {
	resp, err := dbClient.FindAlbumList(context.Background(), &dbService.UserId{UserId: userId})
	if err != nil {
		fmt.Println(err)
		return errors.STATUS_INTERNAL_ERROR
	}

	if !util.IsSucced(resp.GetStatusCode()) {
		fmt.Println("AlbumList Not Found")
		return int(resp.GetStatusCode())
	}

	albumList := &dbService.Albums{}
	err = ptypes.UnmarshalAny(resp.Value, albumList)
	if err != nil {
		fmt.Println(err)
		return errors.STATUS_INTERNAL_ERROR
	}

	albumList.AlbumList = append(albumList.AlbumList, albumId)

	resp, err = dbClient.UpdateAlbumList(context.Background(), &dbService.NewAlbumList{UserId: userId, NewAlbums: albumList})
	if err != nil {
		return errors.STATUS_INTERNAL_ERROR
	}

	if !util.IsSucced(resp.GetStatusCode()) {
		fmt.Println("AlbumList Not Found")
		return int(resp.GetStatusCode())
	}

	return errors.STATUS_SUCCEED_NORETURN
}
*/
