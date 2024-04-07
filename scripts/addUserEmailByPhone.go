package scripts

import (
	"context"
	"hermes/db"

	"go.mongodb.org/mongo-driver/bson"
)

func addUserEmailByPhone() (err error) {

	phoneToEmail := map[string]string{
		"+918800561308": "aryan@amigo.gg",
		"+919910074373": "shashank@amigo.gg",
		"+918426003071": "prashant@amigo.gg",
		"+919571049258": "support@amigo.gg",
		"+919599723760": "muskan@amigo.gg",
		"+917703933820": "tanvi@amigo.gg",
		"+917042072821": "abhishek@amigo.gg",
		"+919990165720": "samrat@amigo.gg",
		"+919911561505": "ishant@amigo.gg",
		"+919089749849": "alou@amigo.gg",
		"+919643099621": "kshitj@amigo.gg",
		"+918700732909": "kshitj1@amigo.gg",
		"+919687029466": "shalin@amigo.gg",
		"+919910606373": "whatsapp@amigo.gg",
		"+919910608373": "business@amigo.gg",
	}

	for key, element := range phoneToEmail {
		_, err := db.UserCollection.UpdateOne(context.Background(), bson.M{"phone": key}, bson.M{"$set": bson.M{
			"email": element,
		}})

		if err != nil {
			panic(err)
		}
	}

	return nil
}
