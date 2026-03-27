/*
 * @Author: 小鱼
 * @Date: 2025-10-21 14:23:13
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-21 16:46:33
 * @FilePath: \passwordManageServer\pkg\middleware\chain.go
 * @Description:
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package middleware

import "net/http"

type Interceptor func(http.HandlerFunc) http.HandlerFunc

func Chain(final http.HandlerFunc, interceptors ...Interceptor) http.HandlerFunc {
	if len(interceptors) == 0 {
		return final
	}
	wrapped := final
	// 倒序包裹，保证先传入的先执行
	for i := len(interceptors) - 1; i >= 0; i-- {
		wrapped = interceptors[i](wrapped)
	}
	return wrapped
}
