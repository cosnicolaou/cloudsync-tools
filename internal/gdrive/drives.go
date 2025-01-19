package gdrive

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func GetCredentials(ctx context.Context) (context.Context, error) {
	creds, err := google.FindDefaultCredentials(ctx, drive.DriveScope)
	fmt.Printf("CREDS: %s.. %v\n", creds, err)
	return ctx, nil
}

/*
 gcloud auth application-default login --scopes=https://www.googleapis.com/auth/drive,https://www.googleapis.com/auth/drive.metadata --client-id-file=$HOME/client_secret_679506199984-uvs3dhtjpsvlgqkerr5t2hp3n8840l78.apps.googleusercontent.com.json
*/

/*
https://developers.google.com/drive/api/guides/search-files

mimeType = 'application/vnd.google-apps.folder'

mimeType = 'application/vnd.google-apps.folder' and name = 'macbook pro 2021 (m1)'

files/name,files/id,files/parents,files/md5Checksum,files/mimeType
*/

func ListFiles(ctx context.Context) error {
	/*
		raw, err := os.ReadFile("/Users/cnicolaou/forward-subject-405104-23d4ef3d795c.json")
		if err != nil {
			return err
		}
		creds, err := google.CredentialsFromJSON(ctx, raw)
		if err != nil {
			return err
		}*/

	driveService, err := drive.NewService(ctx,
		option.WithScopes(
			drive.DriveReadonlyScope,
			drive.DriveScope,
			drive.DriveMetadataScope))

	if err != nil {
		return err

	}

	// includes trashed

	fl, err := driveService.Files.List().Do()
	if err != nil {
		return err
	}
	for _, f := range fl.Files {
		fmt.Printf("f: %v - %v, parents: %v\n", f.Id, f.Name, f.Parents)
		gf, err := driveService.Files.Get(f.Id).Do()
		if err != nil {
			return err
		}
		fmt.Printf("parents: %v\n", gf.Parents)
	}
	return err
}
