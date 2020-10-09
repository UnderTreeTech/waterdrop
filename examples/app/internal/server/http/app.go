/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package http

import (
	"fmt"
	"strconv"
	"time"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/dao"
	"github.com/UnderTreeTech/waterdrop/examples/app/internal/model"
	"github.com/UnderTreeTech/waterdrop/examples/app/internal/utils"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gomodule/redigo/redis"
)

func getAppInfo(c *gin.Context) {
	ctx := c.Request.Context()
	condition := map[string]interface{}{
		"id": 26186061323568199,
	}

	shop, err := dao.GetDao().FindShop(ctx, condition)
	if err != nil {
		log.Error(ctx, "get shop", log.Any("err", err.Error()))
	}

	dao.GetRedis().Do(ctx, "set", "shop_info_"+strconv.Itoa(int(shop.ID)), shop.ShopName)
	dao.GetRedis().Do(ctx, "HSET", "helloworld", "4", "123")
	reply, err := redis.Strings(dao.GetRedis().Do(ctx, "HMGET", "helloworld", "1", "4", "3"))
	fmt.Println(reply, err)

	s := &model.Shop{}
	s.ID = uint64(time.Now().UnixNano())
	s.ShopName = utils.RandomString(16)
	s.Address = utils.RandomString(16)
	s.ProvinceName = utils.RandomString(16)
	s.CityName = utils.RandomString(16)
	s.DistrictName = utils.RandomString(16)
	s.CreatedTime = uint32(utils.GetCurrentUnixTime())
	s.UpdatedTime = uint32(utils.GetCurrentUnixTime())
	if err := dao.GetDao().AddShop(ctx, s); err != nil {
		log.Error(ctx, "add shop", log.Any("err", err))
	}

	setMap := map[string]interface{}{
		"province_name": "上海",
		"city_name":     "上海",
		"district_name": "上海",
	}

	conditionMap := map[string]interface{}{
		"id": []uint64{26186061323568199, 26186061323568200, 26186061323568201},
	}

	if err := dao.GetDao().EditShop(ctx, setMap, conditionMap); err != nil {
		log.Error(ctx, "update shop", log.Any("err", err))
	}

	ctx, err = dao.GetDao().Begin(ctx)
	if err != nil {
		log.Error(ctx, "begin tx", log.Any("err", err))
		return
	}

	shopId := uint64(time.Now().UnixNano())
	s.ID = shopId
	s.ShopName = "小米旗舰店"
	s.CreatedTime = uint32(utils.GetCurrentUnixTime())
	s.UpdatedTime = uint32(utils.GetCurrentUnixTime())
	if err := dao.GetDao().AddShop(ctx, s); err != nil {
		dao.GetDao().Rollback(ctx)
	}

	dao.GetDao().FindShop(ctx, condition)
	if info, err := redis.String(dao.GetRedis().Do(ctx, "get", "shop_info_"+strconv.Itoa(int(shop.ID)))); err != nil {
		log.Info(ctx, "query redis", log.String("shop_info", info))
	}

	conditionMap = map[string]interface{}{"id": shopId}
	if err := dao.GetDao().EditShop(ctx, setMap, conditionMap); err != nil {
		dao.GetDao().Rollback(ctx)
	}

	dao.GetDao().Commit(ctx)

	c.JSON(0, shop)
}

func getSkipUrls(ctx *gin.Context) {
	log.Debugf("test reload log level")
	ctx.JSON(0, "hello")
}

func validateApp(ctx *gin.Context) {
	id := ctx.Params.ByName("id")
	log.Info(ctx.Request.Context(), "", log.String("id", id))

	req := &ValidateReq{}
	if err := ctx.Bind(req); err != nil {
		log.Error(ctx.Request.Context(), "error", log.String("err_msg", err.Error()))
		return
	}

	ctx.JSON(0, Response{Code: 0, Message: "ok", Data: &empty.Empty{}})
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ValidateReq struct {
	Email    string     `protobuf:"bytes,1,opt,name=email,proto3" json:"email" form:"email" validate:"required,email"`
	Name     string     `protobuf:"bytes,2,opt,name=name,proto3" json:"name" form:"name" validate:"required,min=6,max=10"`
	Password string     `protobuf:"bytes,3,opt,name=password,proto3" json:"password" form:"password" validate:"required,min=6,max=10"`
	Sex      int32      `protobuf:"varint,4,opt,name=sex,proto3" json:"sex" form:"sex" validate:"required,gte=0,lte=2"`
	Age      int32      `protobuf:"varint,5,opt,name=age,proto3" json:"age" form:"age" validate:"required,gte=1,lte=60,gtefield=Sex"`
	Addr     []*Address `protobuf:"bytes,6,rep,name=addr,proto3" json:"addr" form:"addr" validate:"required,gt=0,dive"`
}

type Address struct {
	Mobile  string         `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile" form:"mobile" validate:"required,mobile,min=6,max=20"`
	Address string         `protobuf:"bytes,2,opt,name=address,proto3" json:"address" form:"address" validate:"required,max=100"`
	App     *AppReq        `protobuf:"bytes,3,opt,name=app,proto3" json:"app"`
	Reply   *SkipUrlsReply `protobuf:"bytes,4,opt,name=reply,proto3" json:"reply"`
	Resp    []*AppReply    `protobuf:"bytes,5,rep,name=resp,proto3" json:"resp"`
}

type AppReq struct {
	Sappkey string `protobuf:"bytes,1,opt,name=sappkey,proto3" json:"sappkey,omitempty" form:"sappkey" validate:"required"`
}

type SkipUrlsReply struct {
	Urls []string `protobuf:"bytes,1,rep,name=urls,proto3" json:"urls"`
}

type AppReply struct {
	Appkey    string `protobuf:"bytes,1,opt,name=appkey,proto3" json:"app_key"`
	Appsecret string `protobuf:"bytes,2,opt,name=appsecret,proto3" json:"app_secret"`
}
