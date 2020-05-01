package infrastructure

import (
	"github.com/kindaidensan/vote-system/vote"
	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"database/sql"
	"context"
	"errors"
	"log"
	"net"
)

type VoteService struct {
	SqlHandler *SqlHandler
}

func (v *VoteService) Create(c context.Context, r *vote.CreateRequest) (*vote.CreateResponse, error) {
	query := "INSERT INTO  agenda(title, owner, description) VALUES(?, ? ,?);"
	id, err :=  v.SqlHandler.Insert(query, r.Title, r.Owner, r.Description)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	query = "SELECT * FROM agenda WHERE id = " + id + ";"
	rows, err := v.SqlHandler.Query(query)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	for rows.Next() {
		var id string
		var title string
		var owner string
		var description string
		var created string
		var closed sql.NullString
		if err := rows.Scan(&id, &title, &owner, &description, &created, &closed); err != nil {
			return nil, errors.New(err.Error())
		}
		return &vote.CreateResponse {
			Id: id,
			Title: title,
			Description: description,
			Owner: owner,
			Created: created,
		}, nil
	}
	return nil, nil
}

func (v *VoteService) Vote(c context.Context, r *vote.VoteRequest) (*vote.VoteResponse, error) {
	query := "SELECT * FROM agenda WHERE id = " + r.Id + ";"
	rows, err := v.SqlHandler.Query(query)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return &vote.VoteResponse {
			Status: false,
		}, nil
	}
	query = "SELECT * FROM voteduser WHERE agenda_id = " + r.Id + " AND user_id = '" + r.Userid + "';"
	rows, err = v.SqlHandler.Query(query)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return &vote.VoteResponse {
			Status: false,
		}, nil
	}
	query = "INSERT INTO  voteduser(agenda_id, user_id, agree) VALUES(?, ? ,?);"
	agree := 0;
	if r.Agree {
		agree = 1;
	}
	_, err = v.SqlHandler.Insert(query, r.Id, r.Userid, agree)
	if err != nil {
		return &vote.VoteResponse {
			Status: false,
		}, err
	}
	return &vote.VoteResponse {
		Status: true,
	}, nil
}

func (v *VoteService) Delete(c context.Context, r *vote.DeleteRequest) (*vote.DeleteResponse, error) {
	query := "DELETE FROM voteduser WHERE agenda_id = " + r.Id + ";"
	_, err := v.SqlHandler.Query(query)	
	if err != nil {
		return nil, err
	}
	query = "DELETE FROM agenda WHERE id = " + r.Id + " AND owner = '" + r.Userid + "';"
	_, err = v.SqlHandler.Query(query)	
	if err != nil {
		return nil, err
	}
	return &vote.DeleteResponse {
		Status: true,
	}, nil
}

func (v *VoteService) Close(c context.Context, r *vote.CloseRequest) (*vote.CloseResponse, error) {
	query := "UPDATE agenda SET closed_at = current_timestamp WHERE id = " + r.Id + " AND owner = '" + r.Userid + "';"
	_, err := v.SqlHandler.Query(query)	
	if err != nil {
		return nil, err
	}	
	return &vote.CloseResponse {
		Status: true,
	}, nil
}

func (v *VoteService) Get(c context.Context, r *vote.GetRequest) (*vote.GetResponse, error) {
	query := "SELECT a.id, a.title, a.owner, a.description, a.created_at, a.closed_at, COUNT(v.agree = true OR NULL) AS agree, COUNT(v.agree = false OR NULL) AS disagree " +
			 "FROM agenda AS a " +
			 "JOIN voteduser AS v " +
			 "ON a.id = v.agenda_id " +
			 "GROUP BY v.agenda_id;"
	rows, err := v.SqlHandler.Query(query)
	if err != nil {
		return nil, err
	}
	agendas := []*vote.Agenda{}
	for rows.Next() {
		var id string
		var title string 
		var owner string
		var description string
		var created string
		var closed string
		var agree int
		var disagree int
		if err := rows.Scan(&id, &title, &owner, &description, & created, &closed, &agree, &disagree); err != nil {
			return nil, err
		}
		agenda := vote.Agenda {
			Id: id,
			Title: title,
			Owner: owner,
			Description: description,
			Created: created,
			Closed: closed,
			Agree: uint64(agree),
			Disagree: uint64(disagree),
		}
		agendas = append(agendas, &agenda)
	}
	return &vote.GetResponse {
		Agendas: agendas,
	}, nil
}

type Config struct {
	SqlConfig SqlConfig
}

func Start() {
	// config取得
	var config Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		panic(err)
	}

	listenPort, err := net.Listen("tcp", ":8040")
    if err != nil {
        log.Fatalln(err)
    }
	server := grpc.NewServer()
    vote.RegisterVoteServer(server, &VoteService {
		SqlHandler: NewSqlHandler(config.SqlConfig),
	})
	reflection.Register(server)
	err = server.Serve(listenPort)
	if err != nil {
		panic(err)
	}
	log.Println("Listen server")
}