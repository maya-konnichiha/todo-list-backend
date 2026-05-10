package task

import (
	"bytes"
	"encoding/json"
)

// Optional は JSON フィールドの「未指定 / null / 値あり」を区別する汎用型。
// PATCH リクエストで「フィールド未指定 = 変更なし」「null = NULL に更新」「値 = 値で更新」
// の 3 状態を表現するために使う。
//
// UnmarshalJSON が呼ばれた = JSON にフィールドが存在した、なので Present=true。
// 中身が null なら Value=nil、それ以外は T 型の値を Unmarshal して Value にセットする。
type Optional[T any] struct {
	Present bool
	Value   *T
}

// UnmarshalJSON は標準ライブラリの json デコード時に自動で呼ばれる。
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}
