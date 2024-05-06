package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	domainModel "DiTing-Go/domain/model"
	"DiTing-Go/domain/vo/req"
	domainResp "DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	pkgEnum "DiTing-Go/pkg/domain/enum"
	pkgReq "DiTing-Go/pkg/domain/vo/req"
	"DiTing-Go/pkg/domain/vo/resp"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/service/adapter"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

func GetContactListService(uid int64, pageReq pkgReq.PageReq) (pkgResp.ResponseData, error) {
	db := dal.DB
	contact := make([]model.Contact, 0)
	condition := []interface{}{"uid=?", strconv.FormatInt(uid, 10)}
	if pageReq.Cursor != nil && *pageReq.Cursor != "" {
		// 时间戳转时间
		timestamp, err := strconv.ParseInt(*pageReq.Cursor, 10, 64)
		if err != nil {
			global.Logger.Errorf("时间戳转换失败 %s", err)
			return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
		}
		cursor := time.Unix(0, timestamp)
		cursorStr := cursor.Format(time.RFC3339Nano)
		pageReq.Cursor = &cursorStr
	}

	pageResp, err := utils.Paginate(db, pageReq, &contact, "active_time", false, condition...)
	if err != nil {
		global.Logger.Errorf("查询会话列表失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	contactList := pageResp.Data.(*[]model.Contact)

	// 收集会话id
	contactRoomIdList := make([]int64, 0)
	for _, contact := range *contactList {
		contactRoomIdList = append(contactRoomIdList, contact.RoomID)
	}

	// 查询出对应的房间信息
	ctx := context.Background()
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	// 查询房间类型
	roomRList, err := roomQ.Where(room.ID.In(contactRoomIdList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 收集单聊房间的id
	roomFriendIdList := make([]int64, 0)
	// 收集群聊房间的id
	roomGroupIdList := make([]int64, 0)
	for _, room := range roomRList {
		if room.Type == enum.PERSONAL {
			roomFriendIdList = append(roomFriendIdList, room.ID)
		} else if room.Type == enum.GROUP {
			roomGroupIdList = append(roomGroupIdList, room.ID)
		}
	}

	// 查询好友房间信息
	roomFriend := global.Query.RoomFriend
	roomFriendQ := roomFriend.WithContext(ctx)
	roomFriendRList, err := roomFriendQ.Where(roomFriend.RoomID.In(roomFriendIdList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询好友房间失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 查询用户信息
	uidList := make([]int64, 0)
	for _, roomFriend := range roomFriendRList {
		if roomFriend.Uid1 == uid {
			uidList = append(uidList, roomFriend.Uid2)
		} else {
			uidList = append(uidList, roomFriend.Uid1)
		}
	}
	user := global.Query.User
	userQ := user.WithContext(ctx)
	userRList, err := userQ.Where(user.ID.In(uidList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询用户失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 查询群聊房间信息
	roomGroup := global.Query.RoomGroup
	roomGroupQ := roomGroup.WithContext(ctx)
	roomGroupRList, err := roomGroupQ.Where(roomGroup.RoomID.In(roomGroupIdList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询群聊房间失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 查询最后一条消息
	lastMsgIdList := make([]int64, 0)
	for _, contact := range *contactList {
		lastMsgIdList = append(lastMsgIdList, contact.LastMsgID)
	}
	msg := global.Query.Message
	msgQ := msg.WithContext(ctx)
	msgRList, err := msgQ.Where(msg.ID.In(lastMsgIdList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询最后一条消息失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 查询未读消息数
	//TODO:优化
	counts := make([]int, 0)
	for _, contact := range *contactList {
		count, err := msgQ.Where(msg.RoomID.Eq(contact.RoomID), msg.DeleteStatus.Eq(enum.NORMAL), msg.CreateTime.Gt(contact.ReadTime)).Limit(99).Count()
		if err != nil {
			global.Logger.Errorf("统计未读数失败 %s", err)
			return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
		}
		counts = append(counts, int(count))
	}

	// 拼装结果
	contactDaoList := adapter.BuildContactDaoList(*contactList, userRList, msgRList, roomRList, roomFriendRList, roomGroupRList, counts)

	pageResp.Data = contactDaoList
	return pkgResp.SuccessResponseData(pageResp), nil
}

// 获取会话dto
func getContactDto(contact model.Contact) (*dto.ContactDto, error) {
	ctx := context.Background()
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	msg := global.Query.Message
	msgQ := msg.WithContext(ctx)
	contactDto := new(dto.ContactDto)
	contactDto.ID = contact.ID
	// 查询房间类型
	roomR, err := roomQ.Where(room.ID.Eq(contact.RoomID)).First()
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return nil, err
	}
	contactDto.RoomID = roomR.ID
	// 如果是个人会话，名称是对方的昵称
	if roomR.Type == enum.PERSONAL {
		// 查询好友房间信息
		roomFriend := global.Query.RoomFriend
		roomFriendQ := roomFriend.WithContext(ctx)
		roomFriendR, err := roomFriendQ.Where(roomFriend.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询好友房间失败 %s", err)
			return nil, err
		}
		var friendUid int64
		if roomFriendR.Uid1 == contact.UID {
			friendUid = roomFriendR.Uid2
		} else {
			friendUid = roomFriendR.Uid1
		}
		user := global.Query.User
		userQ := user.WithContext(ctx)
		userR, err := userQ.Where(user.ID.Eq(friendUid)).First()
		if err != nil {
			global.Logger.Errorf("查询用户失败 %s", err)
			return nil, err
		}
		contactDto.Name = userR.Name
		contactDto.Avatar = userR.Avatar
		contactDto.LastTime = contact.ActiveTime.UnixNano()
	}
	count, err := msgQ.Where(msg.RoomID.Eq(contact.RoomID), msg.DeleteStatus.Eq(enum.NORMAL), msg.CreateTime.Gt(contact.ReadTime)).Limit(99).Count()
	if err != nil {
		global.Logger.Errorf("统计未读数失败 %s", err)
		return nil, err
	} else if roomR.Type == enum.GROUP {
		// 查询群聊表
		roomGroup := global.Query.RoomGroup
		roomGroupQ := roomGroup.WithContext(ctx)
		roomGroupR, err := roomGroupQ.Where(roomGroup.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询群聊失败 %s", err)
			return nil, err
		}
		contactDto.Name = roomGroupR.Name
		contactDto.Avatar = roomGroupR.Avatar
		contactDto.LastTime = contact.ActiveTime.UnixNano()
		// TODO:热点群聊

	}
	// TODO: 支持多种消息
	msgR, err := msgQ.Where(msg.ID.Eq(contact.LastMsgID)).First()
	message := domainModel.Message(*msgR)
	if err != nil {
		global.Logger.Errorf("查询消息失败 %s", err)
		return nil, err
	}
	contactDto.LastMsg = message.GetContactMsg()
	contactDto.UnreadCount = int32(count)
	// TODO: 群聊
	return contactDto, nil
}

func GetContactDetailService(c *gin.Context) {
	uid := c.GetInt64("uid")
	getMessageListReq := req.GetMessageListReq{}
	if err := c.ShouldBindQuery(&getMessageListReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	roomId := getMessageListReq.RoomId
	cursor, err := timestampToTime(getMessageListReq.Cursor)
	if err != nil {
		global.Logger.Errorf("时间戳转换失败 %s", err)
		resp.ErrorResponse(c, "系统正忙，请稍后再试")
		c.Abort()
		return
	}
	pageRequest := pkgReq.PageReq{
		Cursor:   cursor,
		PageSize: getMessageListReq.PageSize,
	}
	// 更新会话表
	contact := global.Query.Contact
	contactQ := contact.WithContext(context.Background())
	_, err = contactQ.Where(contact.UID.Eq(uid), contact.RoomID.Eq(roomId)).Update(contact.ReadTime, time.Now())
	if err != nil {
		global.Logger.Errorf("更新会话失败 %s", err)
		resp.ErrorResponse(c, "系统正忙，请稍后再试")
		c.Abort()
		return
	}

	// 获取会话详情
	pageResp, err := GetContactDetail(roomId, pageRequest)
	if err != nil {
		global.Logger.Errorf("查询会话详情失败 %s", err)
		resp.ErrorResponse(c, "系统正忙，请稍后再试")
		c.Abort()
		return
	}
	resp.SuccessResponse(c, pageResp)
	return
}

func GetContactDetail(roomID int64, pageRequest pkgReq.PageReq) (*pkgResp.PageResp, error) {
	// 查询消息
	db := dal.DB
	msgs := make([]model.Message, 0)
	// TODO: 抽象成常量
	condition := []interface{}{"room_id=? AND delete_status=?", strconv.FormatInt(roomID, 10), pkgEnum.NORMAL}
	pageResp, err := utils.Paginate(db, pageRequest, &msgs, "create_time", false, condition...)
	if err != nil {
		global.Logger.Errorf("查询消息失败: %s", err.Error())
		return nil, err
	}
	msgList := pageResp.Data.(*[]model.Message)
	userIdMap := make(map[int64]*int64)
	for _, msg := range *msgList {
		if userIdMap[msg.FromUID] == nil {
			userIdMap[msg.FromUID] = &msg.FromUID
		}
	}
	// 转换成列表
	userIdList := make([]int64, 0)
	for _, uid := range userIdMap {
		userIdList = append(userIdList, *uid)
	}
	// 查询用户信息
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	users, err := userQ.Where(user.ID.In(userIdList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询用户失败: %s", err.Error())
		return nil, err
	}
	userMap := make(map[int64]*model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 拼装结果
	pageResp = adapter.BuildMessageRespByMsgAndUser(pageResp, msgList, userMap)
	return pageResp, nil
}
func timestampToTime(timestampStr *string) (*string, error) {
	if timestampStr != nil && *timestampStr != "" {
		// 时间戳转时间
		timestamp, err := strconv.ParseInt(*timestampStr, 10, 64)
		if err != nil {
			global.Logger.Errorf("时间戳转换失败 %s", err)
			return nil, err
		}
		cursor := time.Unix(0, timestamp)
		cursorStr := cursor.Format(time.RFC3339Nano)
		return &cursorStr, nil
	}
	return nil, nil
}

func GetUserInfoBatchService(reqList req.GetUserInfoBatchReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	uids := make([]int64, 0)

	userMap := make(map[int64]*req.UserInfoBatchReqItem)
	for _, item := range reqList.List {
		uids = append(uids, item.Uid)
		userMap[item.Uid] = &item
	}
	users, err := userQ.Where(user.ID.In(uids...)).Find()
	if err != nil {
		global.Logger.Errorf("查询用户失败 %s", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	resultList := make([]domainResp.GetUserInfoBatchResp, 0)
	for _, user := range users {
		resultItem := domainResp.GetUserInfoBatchResp{
			Uid: user.ID,
		}
		if user.UpdateTime.UnixMilli() > userMap[user.ID].LastModifyTime {
			resultItem.Username = user.Name
			resultItem.Avatar = user.Avatar
			resultItem.NeedRefresh = true
		} else {
			resultItem.NeedRefresh = false
		}
		resultList = append(resultList, resultItem)
	}

	return pkgResp.SuccessResponseData(resultList), nil

}
