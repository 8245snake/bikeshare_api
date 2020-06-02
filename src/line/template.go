package main

import (
	"fmt"

	bikeshareapi "github.com/8245snake/bikeshare-client"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	//ColorRegButton 登録ボタンの色
	ColorRegButton = "#00aced"
	//ColorUnregButton 登録ボタンの色
	ColorUnregButton = "#ee0000"
)

//CommandListItem コマンドリストの要素
type CommandListItem struct {
	ActionType        linebot.ActionType
	Label, Data, Text string
}

//TemplateMessageParameter グテンプレートのパラメータ
type TemplateMessageParameter struct {
	Area, Spot, Title, URL, Description, LastUpdate string
	RegButtonVisible                                bool
}

//getLastUpdateTime 「最終更新日時：yyyy/mm/dd hh:mi」の文字列を生成
func getLastUpdateTime(spotinfos ...bikeshareapi.SpotInfo) (lastUpdateTime string) {
	lastUpdateTime = "最終更新日時不明"
	if len(spotinfos) > 0 {
		if len(spotinfos[0].Counts) > 0 {
			lastUpdateTime = fmt.Sprintf("最終更新日時：%s", spotinfos[0].Counts[0].Time.Format("2006/01/02 15:04"))
		}
	}
	return
}

//CreateSpotListBubbleContainer 台数一覧のテンプレート作成
func CreateSpotListBubbleContainer(title, altText string, spotinfos []bikeshareapi.SpotInfo) linebot.BubbleContainer {
	//最終更新日時
	lastUpdateTime := getLastUpdateTime(spotinfos...)
	//ヘッダ
	header := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
	}
	header.Contents = append(header.Contents,
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   title,
			Weight: linebot.FlexTextWeightTypeBold,
			Color:  "#aaaaaa",
			Size:   linebot.FlexTextSizeTypeMd,
			Wrap:   true,
		},
		&linebot.TextComponent{
			Type: linebot.FlexComponentTypeText,
			Text: lastUpdateTime,
			Size: linebot.FlexTextSizeTypeMd,
			Wrap: true,
		})

	//ボディ
	body := linebot.BoxComponent{
		Type:    linebot.FlexComponentTypeBox,
		Layout:  linebot.FlexBoxLayoutTypeVertical,
		Spacing: linebot.FlexComponentSpacingTypeMd,
	}
	for _, info := range spotinfos {
		var listitem string
		if len(info.Counts) > 0 {
			listitem = fmt.Sprintf("[%s-%s] %s (%d台)", info.Area, info.Spot, info.Name, info.Counts[0].Count)
		} else {
			listitem = fmt.Sprintf("[%s-%s] %s (台数不明)", info.Area, info.Spot, info.Name)
		}
		item := CreateListInnerBox(
			listitem,
			ColorRegButton,
			"詳細",
			"グラフ作成中です。\nしばらくお待ち下さい・・・",
			GetPostbackDataForAnalyze(info.Area, info.Spot, 2),
		)
		body.Contents = append(body.Contents,
			&linebot.SeparatorComponent{Type: linebot.FlexComponentTypeSeparator},
			&item,
		)
	}
	body.Contents = append(body.Contents,
		&linebot.SeparatorComponent{Type: linebot.FlexComponentTypeSeparator},
	)

	//フッター
	footer := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
	}
	footer.Contents = append(footer.Contents,
		&linebot.TextComponent{
			Type: linebot.FlexComponentTypeText,
			Text: "「詳細」ボタンをクリックすると時系列グラフを表示します（返信まで2秒程度かかります）",
			Size: linebot.FlexTextSizeTypeXs,
			Wrap: true,
		},
	)
	//メッセージをセット
	container := linebot.BubbleContainer{
		Type:   linebot.FlexContainerTypeBubble,
		Header: &header,
		Body:   &body,
		Footer: &footer,
	}
	return container
}

//CreateListInnerBox リストの中身（台数一覧用）
func CreateListInnerBox(listitem, buttonColor, buttonCaption, postbackText, postbackData string) linebot.BoxComponent {
	item := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeHorizontal,
		Flex:   linebot.IntPtr(1),
	}
	item.Contents = append(item.Contents,
		&linebot.TextComponent{
			Type: linebot.FlexComponentTypeText,
			Text: listitem,
			Size: linebot.FlexTextSizeTypeSm,
			Wrap: true,
			Flex: linebot.IntPtr(9),
		},
		&linebot.ButtonComponent{
			Type:   linebot.FlexComponentTypeButton,
			Margin: linebot.FlexComponentMarginTypeNone,
			Style:  linebot.FlexButtonStyleTypePrimary,
			Height: linebot.FlexButtonHeightTypeSm,
			Flex:   linebot.IntPtr(4),
			Color:  buttonColor,
			Action: linebot.NewPostbackAction(buttonCaption, postbackData, "", postbackText),
		},
	)
	return item
}

//CreateListInnerBoxHalf リストの中身（テキストとボタンの幅が１：１）時刻設定用
func CreateListInnerBoxHalf(listitem, buttonColor, buttonCaption, postbackText, postbackData string) linebot.BoxComponent {
	var action linebot.TemplateAction
	//postbackDataの中身を見てボタンの挙動を変化させる
	switch command := ParsePostbackData(postbackData); command.Mode {
	case PostBackCommandModeReg:
		action = linebot.NewDatetimePickerAction(buttonCaption, postbackData, "time", "", "", "")
	case PostBackCommandModeUnreg:
		action = linebot.NewPostbackAction(buttonCaption, postbackData, "", postbackText)
	}
	item := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeHorizontal,
		Flex:   linebot.IntPtr(1),
	}
	item.Contents = append(item.Contents,
		&linebot.TextComponent{
			Type: linebot.FlexComponentTypeText,
			Text: listitem,
			Size: linebot.FlexTextSizeTypeXxl,
			Wrap: true,
			Flex: linebot.IntPtr(1),
		},
		&linebot.ButtonComponent{
			Type:   linebot.FlexComponentTypeButton,
			Margin: linebot.FlexComponentMarginTypeNone,
			Style:  linebot.FlexButtonStyleTypePrimary,
			Height: linebot.FlexButtonHeightTypeSm,
			Flex:   linebot.IntPtr(1),
			Color:  buttonColor,
			Action: action,
		},
	)
	return item
}

//CreateSpotListCarouselContainer 件数が多いとき用のテンプレート
func CreateSpotListCarouselContainer(title, altText string, spotinfos []bikeshareapi.SpotInfo) linebot.CarouselContainer {
	contents := CreateSpotListBubbleContainer(title, altText, spotinfos)
	container := linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: []*linebot.BubbleContainer{&contents},
	}
	return container
}

//CreateAnalysisBubbleContainer グラフのコンテナ作成
func CreateAnalysisBubbleContainer(param TemplateMessageParameter) linebot.BubbleContainer {
	var label, text, color string
	label = "お気に入りに登録する"
	text = "お気に入りに登録しています"
	color = ColorRegButton
	postbackdataFavList := GetPostbackDataForFovarite(param.Area, param.Spot, PostBackCommandModeReg)
	postbackdataDatePicker := GetPostbackDataForDateAnalyze(param.Area, param.Spot)

	//ヘッダ
	header := linebot.BoxComponent{
		Type:    linebot.FlexComponentTypeBox,
		Layout:  linebot.FlexBoxLayoutTypeVertical,
		Spacing: linebot.FlexComponentSpacingTypeXs,
	}
	header.Contents = append(header.Contents,
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   param.Title,
			Weight: linebot.FlexTextWeightTypeBold,
			Color:  "#222222",
			Size:   linebot.FlexTextSizeTypeMd,
			Align:  linebot.FlexComponentAlignTypeStart,
			Wrap:   true,
		},
	)
	//ヒーロー
	hero := linebot.ImageComponent{
		Type:        linebot.FlexComponentTypeImage,
		URL:         param.URL,
		Size:        linebot.FlexImageSizeTypeFull,
		AspectRatio: linebot.FlexImageAspectRatioType4to3,
		AspectMode:  linebot.FlexImageAspectModeTypeCover,
	}

	//ボディ
	inner := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
		Flex:   linebot.IntPtr(1),
	}
	//説明を省略するパターンのときもある
	if param.Description != "" {
		inner.Contents = append(inner.Contents,
			&linebot.TextComponent{ //説明
				Type:    linebot.FlexComponentTypeText,
				Text:    param.Description,
				Gravity: linebot.FlexComponentGravityTypeTop,
				Size:    linebot.FlexTextSizeTypeXs,
				Flex:    linebot.IntPtr(1),
				Wrap:    true,
			},
			&linebot.SeparatorComponent{},
		)
	}
	//最終更新日時がないときもあるため
	if param.LastUpdate != "" {
		inner.Contents = append(inner.Contents,
			&linebot.TextComponent{ //最終更新日時
				Type:    linebot.FlexComponentTypeText,
				Text:    param.LastUpdate,
				Align:   linebot.FlexComponentAlignTypeEnd,
				Gravity: linebot.FlexComponentGravityTypeBottom,
				Size:    linebot.FlexTextSizeTypeXs,
				Flex:    linebot.IntPtr(2),
			},
		)
	}
	//お気に入り未登録ならボタン表示
	if param.RegButtonVisible {
		inner.Contents = append(inner.Contents,
			&linebot.ButtonComponent{
				Type:   linebot.FlexComponentTypeButton,
				Margin: linebot.FlexComponentMarginTypeNone,
				Style:  linebot.FlexButtonStyleTypePrimary,
				Height: linebot.FlexButtonHeightTypeSm,
				Flex:   linebot.IntPtr(1),
				Color:  color,
				Action: linebot.NewPostbackAction(label, postbackdataFavList, "", text),
			},
			&linebot.SeparatorComponent{
				Type:   linebot.FlexComponentTypeSeparator,
				Color:  "#FFFFFF",
				Margin: linebot.FlexComponentMarginTypeNone,
			},
		)
	}
	inner.Contents = append(inner.Contents,
		&linebot.ButtonComponent{
			Type:   linebot.FlexComponentTypeButton,
			Margin: linebot.FlexComponentMarginTypeNone,
			Style:  linebot.FlexButtonStyleTypePrimary,
			Height: linebot.FlexButtonHeightTypeSm,
			Flex:   linebot.IntPtr(1),
			Color:  color,
			Action: linebot.NewDatetimePickerAction("別の日のグラフを表示する", postbackdataDatePicker, "date", "", "2020-12-31", "2019-06-01"),
		},
	)
	body := linebot.BoxComponent{
		Type:    linebot.FlexComponentTypeBox,
		Layout:  linebot.FlexBoxLayoutTypeVertical,
		Spacing: linebot.FlexComponentSpacingTypeMd,
	}
	body.Contents = append(body.Contents, &inner)

	//メッセージをセット
	container := linebot.BubbleContainer{
		Type:   linebot.FlexContainerTypeBubble,
		Header: &header,
		Hero:   &hero,
		Body:   &body,
	}
	return container
}

//CreateCommandListBubbleContainer コマンドの一覧画面（履歴一覧やコマンド一覧に使用する）
func CreateCommandListBubbleContainer(title string, commands []CommandListItem) linebot.BubbleContainer {
	//ボディ
	body := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
	}
	for _, item := range commands {
		var action linebot.TemplateAction
		switch item.ActionType {
		case linebot.ActionTypePostback:
			action = linebot.NewPostbackAction(item.Label, item.Data, "", item.Text)
		case linebot.ActionTypeMessage:
			action = linebot.NewMessageAction(item.Label, item.Data)
		}

		body.Contents = append(body.Contents,
			&linebot.ButtonComponent{
				Type:   linebot.FlexComponentTypeButton,
				Margin: linebot.FlexComponentMarginTypeNone,
				Style:  linebot.FlexButtonStyleTypeSecondary,
				Height: linebot.FlexButtonHeightTypeSm,
				Action: action,
			},
			&linebot.SeparatorComponent{
				Type:   linebot.FlexComponentTypeSeparator,
				Color:  "#FFFFFF",
				Margin: linebot.FlexComponentMarginTypeNone,
			},
		)
	}
	//メッセージをセット
	container := linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		// Header: &header,
		Body: &body,
	}
	return container
}

//CreateConfigBubbleContainer 設定画面作成
func CreateConfigBubbleContainer(user *bikeshareapi.Users) linebot.BubbleContainer {
	//ボディ
	body := linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
	}
	body.Contents = append(body.Contents,
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   "ユーザー設定",
			Align:  linebot.FlexComponentAlignTypeCenter,
			Weight: linebot.FlexTextWeightTypeBold,
			Color:  "#1DB446",
			Size:   linebot.FlexTextSizeTypeXl,
		},
		&linebot.SeparatorComponent{
			Margin: linebot.FlexComponentMarginTypeMd,
		},
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   "お気に入り登録されたスポット",
			Color:  "#aaaaaa",
			Size:   linebot.FlexTextSizeTypeXs,
			Margin: linebot.FlexComponentMarginTypeXl,
		},
	)
	for _, code := range user.Favorites {
		area, spot := SplitAreaSpot(code)
		item := CreateListInnerBox(
			fmt.Sprintf("[%s] %s", code, GetPlaceNameByCode(code)),
			ColorUnregButton,
			"削除",
			"お気に入りから削除しています",
			GetPostbackDataForFovarite(area, spot, PostBackCommandModeUnreg),
		)
		body.Contents = append(body.Contents,
			&item,
			&linebot.SeparatorComponent{
				Color: "#ffffff",
			},
		)
	}

	body.Contents = append(body.Contents,
		&linebot.SeparatorComponent{
			Margin: linebot.FlexComponentMarginTypeMd,
		},
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   "お気に入り登録したスポットの通知時刻の設定（2件まで設定できます）",
			Color:  "#aaaaaa",
			Size:   linebot.FlexTextSizeTypeXs,
			Margin: linebot.FlexComponentMarginTypeXl,
			Wrap:   true,
		},
	)
	for i := 0; i < MaxNotifyTimes; i++ {
		if len(user.Notifies) > i {
			item := CreateListInnerBoxHalf(
				user.Notifies[i],
				ColorUnregButton,
				"削除",
				"削除しています",
				GetPostbackDataForNotify(PostBackCommandModeUnreg, user.Notifies[i]),
			)
			body.Contents = append(body.Contents,
				&item,
				&linebot.SeparatorComponent{
					Color: "#ffffff",
				},
			)
		} else {
			item := CreateListInnerBoxHalf(
				"未登録",
				ColorRegButton,
				"新規登録",
				" 登録しています",
				GetPostbackDataForNotify(PostBackCommandModeReg, ""),
			)
			body.Contents = append(body.Contents,
				&item,
				&linebot.SeparatorComponent{
					Color: "#ffffff",
				},
			)
		}
	}

	//メッセージをセット
	container := linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Body: &body,
	}
	return container
}
