package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	brandtypes "github.com/galaxies-labs/galaxy/x/brand/types"
	"github.com/galaxies-labs/galaxy/x/nft/types"
)

func (suite *KeeperTestSuite) TestCreateClass() {

	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb")

	desc := types.NewClassDescription("name", "", "", "")
	//invalid arguments
	//details message in sdk.BasicValidate
	tests := []struct {
		msg         *types.MsgCreateClass
		expectError error
	}{
		{types.NewMsgCreateClass("", classIDA, ownerA.String(), 10_000, desc), brandtypes.ErrInvalidBrandID},
		{types.NewMsgCreateClass(brandIDA, "", ownerA.String(), 10_000, desc), types.ErrInvalidClassID},
		{types.NewMsgCreateClass(brandIDA, classIDA, ownerA.String(), 10_001, desc), types.ErrInvalidFeeBasisPoints},
		{types.NewMsgCreateClass(brandIDA, classIDA, ownerA.String(), 0, types.NewClassDescription("", "", "", "")), types.ErrInvalidClassDescription},
		{types.NewMsgCreateClass(brandIDA, classIDA, "", 10_000, desc), nil},
	}

	for _, test := range tests {
		res, err := msgServer.CreateClass(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)

		if test.expectError != nil {
			fmt.Println(err, test.expectError)
			suite.Require().Equal(err, test.expectError)
		}
	}

	//owner should be same for brand
	classes := []struct {
		b string
		c string
		o sdk.AccAddress
	}{
		{brandIDA, classIDA, ownerA}, {brandIDA, classIDB, ownerA}, {brandIDA, classIDC, ownerA},
		{brandIDB, classIDA, ownerB}, {brandIDB, classIDB, ownerB},
	}

	for _, d := range classes {
		hasBrand := app.BrandKeeper.HasBrand(ctx, d.b)

		msg := types.NewMsgCreateClass(d.b, d.c, d.o.String(), 10_000, desc)

		if !hasBrand {
			_, err := msgServer.CreateClass(wrapCtx, msg)
			suite.Require().Equal(err, brandtypes.ErrNotFoundBrand)

			suite.Require().NoError(
				app.BrandKeeper.SetBrand(ctx, brandtypes.NewBrand(d.b, d.o, brandtypes.NewBrandDescription("name", "", ""))),
			)
		}

		msg.Creator = sdk.AccAddress("randomaddress").String()
		_, err := msgServer.CreateClass(wrapCtx, msg)
		suite.Require().Equal(err, types.ErrUnauthorized)

		_, err = msgServer.CreateClass(wrapCtx, msg)
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) TestEditClass() {
	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb")

	desc := types.NewClassDescription("name", "", "", "")
	//invalid arguments
	//details message in sdk.BasicValidate
	tests := []struct {
		msg         *types.MsgEditClass
		expectError error
	}{
		{types.NewMsgEditClass("", classIDA, ownerA.String(), 10_000, desc), brandtypes.ErrInvalidBrandID},
		{types.NewMsgEditClass(brandIDA, "", ownerA.String(), 10_000, desc), types.ErrInvalidClassID},
		{types.NewMsgEditClass(brandIDA, classIDA, ownerA.String(), 10_001, desc), types.ErrInvalidFeeBasisPoints},
		{types.NewMsgEditClass(brandIDA, classIDA, ownerA.String(), 0, types.NewClassDescription("", "", "", "")), types.ErrInvalidClassDescription},
		{types.NewMsgEditClass(brandIDA, classIDA, "", 10_000, desc), nil},
	}

	for _, test := range tests {
		res, err := msgServer.EditClass(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)
		if test.expectError != nil {
			suite.Require().Equal(err, test.expectError)
		}
	}

	//owner should be same for brand
	classes := []struct {
		b string
		c string
		o sdk.AccAddress
	}{
		{brandIDA, classIDA, ownerA}, {brandIDA, classIDB, ownerA}, {brandIDA, classIDC, ownerA},
		{brandIDB, classIDA, ownerB}, {brandIDB, classIDB, ownerB},
	}

	for _, d := range classes {
		hasBrand := app.BrandKeeper.HasBrand(ctx, d.b)

		msg := types.NewMsgEditClass(d.b, d.c, d.o.String(), 10_000, desc)

		if !hasBrand {
			_, err := msgServer.EditClass(wrapCtx, msg)
			suite.Require().Equal(err, brandtypes.ErrNotFoundBrand)

			suite.Require().NoError(
				app.BrandKeeper.SetBrand(ctx, brandtypes.NewBrand(d.b, d.o, brandtypes.NewBrandDescription("name", "", ""))),
			)
		}

		_, err := msgServer.EditClass(wrapCtx, msg)
		suite.Require().Equal(err, types.ErrNotFoundClass)

		savedClass := types.NewClass(msg.BrandId, msg.Id, msg.FeeBasisPoints, msg.Description)
		suite.Require().NoError(app.NFTKeeper.SaveClass(ctx, savedClass))

		class, exist := app.NFTKeeper.GetClass(ctx, msg.BrandId, msg.Id)
		suite.Require().True(exist)
		suite.Require().Equal(class, savedClass)

		savedClass = types.NewClass(msg.BrandId, msg.Id, 123, types.NewClassDescription("newname", "details", "https://galaxy", "ipfs://image"))
		msg.Description = savedClass.Description
		msg.FeeBasisPoints = savedClass.FeeBasisPoints
		_, err = msgServer.EditClass(wrapCtx, msg)
		suite.Require().NoError(err)

		suite.Require().NotEqual(class, savedClass)
		class, exist = app.NFTKeeper.GetClass(ctx, msg.BrandId, msg.Id)
		suite.Require().True(exist)
		suite.Require().Equal(class, savedClass)

		msg.Editor = sdk.AccAddress("randomaddress").String()
		_, err = msgServer.EditClass(wrapCtx, msg)
		suite.Require().Equal(err, types.ErrUnauthorized)
	}
}

func (suite *KeeperTestSuite) TestMintNFT() {
	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB, ownerC, recipient := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb"), sdk.AccAddress("ownerc"), sdk.AccAddress("recipient")

	//invalid arguments
	tests := []struct {
		msg         *types.MsgMintNFT
		expectError error
	}{
		{types.NewMsgMintNFT("", classIDA, "ipfs://nft", "", ownerA.String(), recipient.String()), brandtypes.ErrInvalidBrandID},
		{types.NewMsgMintNFT(brandIDA, "", "ipfs://nft", "", ownerA.String(), recipient.String()), types.ErrInvalidClassID},
		{types.NewMsgMintNFT(brandIDA, classIDA, "", "", ownerA.String(), recipient.String()), types.ErrInvalidNFTUri},
		{types.NewMsgMintNFT(brandIDA, classIDA, "ipfs://nft", "", "", recipient.String()), nil},
		{types.NewMsgMintNFT(brandIDA, classIDA, "ipfs://nft", "", ownerA.String(), ""), nil},
	}

	for _, test := range tests {
		res, err := msgServer.MintNFT(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)
		if test.expectError != nil {
			suite.Require().Equal(err, test.expectError)
		}
	}

	nftData := []struct {
		num int
		b   string
		c   string
		o   sdk.AccAddress
	}{
		{3, brandIDA, classIDA, ownerA}, {10, brandIDA, classIDA, ownerB}, {2, brandIDA, classIDB, ownerA}, {2, brandIDA, classIDC, ownerA},
		{5, brandIDB, classIDA, ownerB}, {3, brandIDB, classIDB, ownerB}, {10, brandIDB, classIDB, ownerC},
	}

	for _, d := range nftData {
		for i := 0; i < d.num; i++ {
			hasBrand := app.BrandKeeper.HasBrand(ctx, d.b)

			nft, err := app.NFTKeeper.GenNFT(ctx, d.b, d.c, "ipfs://nft", "")
			suite.Require().NoError(err)

			msg := types.NewMsgMintNFT(nft.BrandId, nft.ClassId, "ipfs://nft", "", recipient.String(), recipient.String())

			if !hasBrand {
				_, err = msgServer.MintNFT(wrapCtx, msg)
				suite.Require().Equal(err, brandtypes.ErrNotFoundBrand)

				suite.Require().NoError(
					app.BrandKeeper.SetBrand(ctx, brandtypes.NewBrand(nft.BrandId, recipient, brandtypes.NewBrandDescription("name", "", ""))),
				)
			}

			hasClass := app.NFTKeeper.HasClass(ctx, nft.BrandId, nft.ClassId)
			if !hasClass {
				_, err = msgServer.MintNFT(wrapCtx, msg)
				suite.Require().Equal(err, types.ErrNotFoundClass)

				suite.Require().NoError(
					app.NFTKeeper.SaveClass(ctx, types.NewClass(nft.BrandId, nft.ClassId, 0, types.NewClassDescription("", "", "", ""))),
				)
			}

			nmsg := *msg
			nmsg.Minter = sdk.AccAddress("randomaddress").String()
			_, err = msgServer.MintNFT(wrapCtx, &nmsg)
			suite.Require().Equal(err, types.ErrUnauthorized)

			res, err := msgServer.MintNFT(wrapCtx, msg)
			suite.Require().NoError(err)
			suite.Require().NotZero(res.Id)
		}
	}
}

func (suite *KeeperTestSuite) TestUpdateNFT() {
	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB, ownerC := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb"), sdk.AccAddress("ownerc")

	//invalid arguments
	tests := []struct {
		msg         *types.MsgUpdateNFT
		expectError error
	}{
		{types.NewMsgUpdateNFT("", classIDA, 1, "", ownerA.String()), brandtypes.ErrInvalidBrandID},
		{types.NewMsgUpdateNFT(brandIDA, "", 1, "", ownerA.String()), types.ErrInvalidClassID},
		{types.NewMsgUpdateNFT(brandIDA, classIDA, 0, "", ownerA.String()), types.ErrInvalidNFTID},
		{types.NewMsgUpdateNFT(brandIDA, classIDA, 1, "", ""), nil},
	}

	for _, test := range tests {
		res, err := msgServer.UpdateNFT(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)
		if test.expectError != nil {
			suite.Require().Equal(err, test.expectError)
		}
	}

	//ignore set brand first
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDB, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDC, 0, types.NewClassDescription("", "", "", "")))

	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDB, 0, types.NewClassDescription("", "", "", "")))

	nftData := []struct {
		num int
		b   string
		c   string
		o   sdk.AccAddress
	}{
		{3, brandIDA, classIDA, ownerA}, {10, brandIDA, classIDA, ownerB}, {2, brandIDA, classIDB, ownerA}, {2, brandIDA, classIDC, ownerA},
		{5, brandIDB, classIDA, ownerB}, {3, brandIDB, classIDB, ownerB}, {10, brandIDB, classIDB, ownerC},
	}

	//mint nfts
	for _, d := range nftData {
		for i := 0; i < d.num; i++ {
			nft, err := app.NFTKeeper.GenNFT(ctx, d.b, d.c, "ipfs://nft", "")
			suite.Require().NoError(err)

			msg := types.NewMsgUpdateNFT(nft.BrandId, nft.ClassId, nft.Id, "", d.o.String())
			_, err = msgServer.UpdateNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrNotFoundNFT)

			suite.Require().NoError(app.NFTKeeper.MintNFT(ctx, nft, d.o))

			msg.VarUri = "ipfs://customer_uri"
			_, err = msgServer.UpdateNFT(wrapCtx, msg)
			suite.Require().NoError(err)

			msg.Sender = sdk.AccAddress("randomowner").String()
			_, err = msgServer.UpdateNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrUnauthorized)
		}
	}
}

func (suite *KeeperTestSuite) TestBurnNFT() {
	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB, ownerC := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb"), sdk.AccAddress("ownerc")

	//invalid arguments
	tests := []struct {
		msg         *types.MsgBurnNFT
		expectError error
	}{
		{types.NewMsgBurnNFT("", classIDA, 1, ownerA.String()), brandtypes.ErrInvalidBrandID},
		{types.NewMsgBurnNFT(brandIDA, "", 1, ownerA.String()), types.ErrInvalidClassID},
		{types.NewMsgBurnNFT(brandIDA, classIDA, 0, ownerA.String()), types.ErrInvalidNFTID},
		{types.NewMsgBurnNFT(brandIDA, classIDA, 1, ""), nil},
	}

	for _, test := range tests {
		res, err := msgServer.BurnNFT(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)
		if test.expectError != nil {
			suite.Require().Equal(err, test.expectError)
		}
	}

	//ignore set brand first
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDB, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDC, 0, types.NewClassDescription("", "", "", "")))

	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDB, 0, types.NewClassDescription("", "", "", "")))

	nftData := []struct {
		num int
		b   string
		c   string
		o   sdk.AccAddress
	}{
		{3, brandIDA, classIDA, ownerA}, {10, brandIDA, classIDA, ownerB}, {2, brandIDA, classIDB, ownerA}, {2, brandIDA, classIDC, ownerA},
		{5, brandIDB, classIDA, ownerB}, {3, brandIDB, classIDB, ownerB}, {10, brandIDB, classIDB, ownerC},
	}

	for _, d := range nftData {
		for i := 0; i < d.num; i++ {
			nft, err := app.NFTKeeper.GenNFT(ctx, d.b, d.c, "ipfs://nft", "")
			suite.Require().NoError(err)

			msg := types.NewMsgBurnNFT(nft.BrandId, nft.ClassId, nft.Id, d.o.String())
			_, err = msgServer.BurnNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrNotFoundNFT)

			suite.Require().NoError(app.NFTKeeper.MintNFT(ctx, nft, d.o))

			msg.Sender = sdk.AccAddress("randomowner").String()
			_, err = msgServer.BurnNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrUnauthorized)

			msg.Sender = d.o.String()
			_, err = msgServer.BurnNFT(wrapCtx, msg)
			suite.Require().NoError(err)

			suite.Require().True(
				app.NFTKeeper.BurnedNFT(ctx, nft.BrandId, nft.ClassId, nft.Id),
			)
		}
	}
}

func (suite *KeeperTestSuite) TestTransferNFT() {
	app, ctx, msgServer := suite.app, suite.ctx, suite.msgServer
	wrapCtx := sdk.WrapSDKContext(ctx)

	brandIDA, brandIDB := "brandIDA", "brandIDB"
	classIDA, classIDB, classIDC := "classIDA", "classIDB", "classIDC"
	ownerA, ownerB, ownerC, recipient := sdk.AccAddress("ownera"), sdk.AccAddress("ownerb"), sdk.AccAddress("ownerc"), sdk.AccAddress("recipient")

	//invalid arguments
	tests := []struct {
		msg         *types.MsgTransferNFT
		expectError error
	}{
		{types.NewMsgTransferNFT("", classIDA, 1, ownerA.String(), recipient.String()), brandtypes.ErrInvalidBrandID},
		{types.NewMsgTransferNFT(brandIDA, "", 1, ownerA.String(), recipient.String()), types.ErrInvalidClassID},
		{types.NewMsgTransferNFT(brandIDA, classIDA, 0, ownerA.String(), recipient.String()), types.ErrInvalidNFTID},
		{types.NewMsgTransferNFT(brandIDA, classIDA, 1, "", recipient.String()), nil},
		{types.NewMsgTransferNFT(brandIDA, classIDA, 1, ownerA.String(), ""), nil},
	}

	for _, test := range tests {
		res, err := msgServer.TransferNFT(wrapCtx, test.msg)
		suite.Require().Error(err)
		suite.Require().Nil(res)
		if test.expectError != nil {
			suite.Require().Equal(err, test.expectError)
		}
	}

	//ignore set brand first
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDB, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDA, classIDC, 0, types.NewClassDescription("", "", "", "")))

	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDA, 0, types.NewClassDescription("", "", "", "")))
	app.NFTKeeper.SaveClass(ctx, types.NewClass(brandIDB, classIDB, 0, types.NewClassDescription("", "", "", "")))

	nftData := []struct {
		num int
		b   string
		c   string
		o   sdk.AccAddress
	}{
		{3, brandIDA, classIDA, ownerA}, {10, brandIDA, classIDA, ownerB}, {2, brandIDA, classIDB, ownerA}, {2, brandIDA, classIDC, ownerA},
		{5, brandIDB, classIDA, ownerB}, {3, brandIDB, classIDB, ownerB}, {10, brandIDB, classIDB, ownerC},
	}

	for _, d := range nftData {
		for i := 0; i < d.num; i++ {
			nft, err := app.NFTKeeper.GenNFT(ctx, d.b, d.c, "ipfs://nft", "")
			suite.Require().NoError(err)

			msg := types.NewMsgTransferNFT(nft.BrandId, nft.ClassId, nft.Id, d.o.String(), recipient.String())
			_, err = msgServer.TransferNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrNotFoundNFT)

			suite.Require().NoError(app.NFTKeeper.MintNFT(ctx, nft, d.o))

			msg.Sender = sdk.AccAddress("randomowner").String()
			_, err = msgServer.TransferNFT(wrapCtx, msg)
			suite.Require().Equal(err, types.ErrUnauthorized)

			suite.Require().NotEqual(d.o, recipient)
			suite.Require().Equal(d.o, app.NFTKeeper.GetOwner(ctx, nft.BrandId, nft.ClassId, nft.Id))

			msg.Sender = d.o.String()
			_, err = msgServer.TransferNFT(wrapCtx, msg)
			suite.Require().NoError(err)

			suite.Require().Equal(recipient, app.NFTKeeper.GetOwner(ctx, nft.BrandId, nft.ClassId, nft.Id))

			//check transfer same receipient is passed
			msg.Sender = msg.Recipient
			_, err = msgServer.TransferNFT(wrapCtx, msg)
			suite.Require().NoError(err)
		}
	}
}
