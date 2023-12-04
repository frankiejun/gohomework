package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao/article"
)

//go:generate mockgen -source=./article_reader.go -package=repomocks -destination=mocks/article_reader.mock.go ArticleReaderRepository
type ArticleReaderRepository interface {
	Save(ctx context.Context, art domain.Article) error
}

func NewCachedArticleReaderRepository(dao article.ArticleReaderDAO) ArticleReaderRepository {
	return &CachedArticleReaderRepository{
		dao: dao,
	}
}

type CachedArticleReaderRepository struct {
	dao article.ArticleReaderDAO
}

func (repo *CachedArticleReaderRepository) Save(ctx context.Context, art domain.Article) error {
	return repo.dao.Upsert(ctx, repo.toEntity(art))
}

// toEntity 理论上来说各个 repository 都有差异，所以复制粘贴也没关系。
// 做成一个包方法也可以，看你喜好。
func (repo *CachedArticleReaderRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
