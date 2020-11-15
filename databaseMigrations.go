package main

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

func migrateDb() error {
	startWritingToDb()
	defer finishWritingToDb()
	m, err := migrator.New(
		migrator.Migrations(
			&migrator.Migration{
				Name: "00001",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`
					create table posts (path text not null primary key, content text, published text, updated text, blog text not null, section text);
					create table post_parameters (id integer primary key autoincrement, path text not null, parameter text not null, value text);
					create index index_pp_path on post_parameters (path);
					create trigger after delete on posts begin delete from post_parameters where path = old.path; end;
					create table indieauthauth (time text not null, code text not null, me text not null, client text not null, redirect text not null, scope text not null);
					create table indieauthtoken (time text not null, token text not null, me text not null, client text not null, scope text not null);
					create index index_iat_token on indieauthtoken (token);
					create table autocert (key text not null primary key, data blob not null, created text not null);
					create table activitypub_followers (blog text not null, follower text not null, inbox text not null, primary key (blog, follower));
					create table webmentions (id integer primary key autoincrement, source text not null, target text not null, created integer not null, status text not null default "new", title text, content text, author text, type text, unique(source, target));
					create index index_wm_target on webmentions (target);
					`)
					return err
				},
			},
			&migrator.Migration{
				Name: "00002",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`
					drop table autocert;
					`)
					return err
				},
			},
			&migrator.Migration{
				Name: "00003",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`
					drop trigger AFTER;
					create trigger trigger_posts_delete_pp after delete on posts begin delete from post_parameters where path = old.path; end;
					`)
					return err
				},
			},
			&migrator.Migration{
				Name: "00004",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`
					create view view_posts_with_title as select id, path, title, content, published, updated, blog, section from (select p.rowid as id, p.path as path, pp.value as title, content, published, updated, blog, section from posts p left outer join post_parameters pp on p.path = pp.path where pp.parameter = 'title');
					create virtual table posts_fts using fts5(path unindexed, title, content, published unindexed, updated unindexed, blog unindexed, section unindexed, content=view_posts_with_title, content_rowid=id);
					insert into posts_fts(posts_fts) values ('rebuild');
					`)
					return err
				},
			},
		),
	)
	if err != nil {
		return err
	}
	if err := m.Migrate(appDb); err != nil {
		return err
	}
	return nil
}
