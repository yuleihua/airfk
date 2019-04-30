// Copyright 2018 The huayulei_2003@hotmail.com Authors
// This file is part of the airfk library.
//
// The airfk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The airfk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the airfk library. If not, see <http://www.gnu.org/licenses/>.
package common

type Result struct {
	Id        string `json:"name"`
	BeginTime int64  `json:"begin_time"`
	EndTime   int64  `json:"end_time"`
	ErrorMsg  string `json:"error"`
	Extra     []byte `json:"output"`
}

func NewResult(id string, begin int64) *Result {
	return &Result{
		Id:        id,
		BeginTime: begin,
	}
}

func NewResultWithEnd(id string, begin, end int64, msg string, extra []byte) *Result {
	return &Result{
		Id:        id,
		BeginTime: begin,
		EndTime:   end,
		ErrorMsg:  msg,
		Extra:     extra,
	}
}

func (r *Result) Set(end int64, msg string, extra []byte) {
	if r != nil {
		r.EndTime = end
		r.ErrorMsg = msg
		r.Extra = extra
	}
}
