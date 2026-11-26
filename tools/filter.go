package tools

import "simple-userfollow-graphql/graph/model"

func FilterToFilterQuery(f []*model.Filter) []*model.FilterQuery {
	var fq []*model.FilterQuery
	for _, v := range f {
		fq = append(fq, &model.FilterQuery{
			Field:  v.Field,
			Value:  v.Value,
			Type:   v.Type,
			Clause: v.Clause,
		})
	}

	return fq
}
