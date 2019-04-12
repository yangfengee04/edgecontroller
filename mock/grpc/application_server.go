// Copyright 2019 Smart-Edge.com, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/satori/go.uuid"
	"github.com/smartedgemec/controller-ce/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type applicationServer struct {
	// maps of application ID to application
	containerApps map[string]*pb.Application
	vmApps        map[string]*pb.Application

	// reference to policy server
	policyServer *applicationPolicyServer
}

func newApplicationServer() *applicationServer {
	return &applicationServer{
		containerApps: make(map[string]*pb.Application),
		vmApps:        make(map[string]*pb.Application),
	}
}

func (s *applicationServer) DeployContainer(
	ctx context.Context,
	containerApp *pb.Application,
) (*pb.ApplicationID, error) {
	id := uuid.NewV4().String()
	s.containerApps[id] = containerApp
	containerApp.Id = id
	containerApp.Status = pb.LifecycleStatus_STOPPED
	s.policyServer.policies[id] = defaultPolicy(id)

	return &pb.ApplicationID{Id: id}, nil
}

func (s *applicationServer) DeployVM(
	ctx context.Context,
	vmApp *pb.Application,
) (*pb.ApplicationID, error) {
	id := uuid.NewV4().String()
	s.vmApps[id] = vmApp
	vmApp.Id = id
	vmApp.Status = pb.LifecycleStatus_STOPPED

	return &pb.ApplicationID{Id: id}, nil
}

func (s *applicationServer) GetAll(
	context.Context,
	*empty.Empty,
) (*pb.Applications, error) {
	var apps []*pb.Application

	for _, containerApp := range s.containerApps {
		apps = append(apps, containerApp)
	}

	for _, vmApp := range s.vmApps {
		apps = append(apps, vmApp)
	}

	return &pb.Applications{
		Applications: apps,
	}, nil
}

func (s *applicationServer) Get(
	ctx context.Context,
	id *pb.ApplicationID,
) (*pb.Application, error) {
	if containerApp, ok := s.containerApps[id.Id]; ok {
		return containerApp, nil
	}

	if vmApp, ok := s.vmApps[id.Id]; ok {
		return vmApp, nil
	}

	return nil, status.Errorf(codes.NotFound, "Application %s not found", id.Id)
}

func (s *applicationServer) Redeploy(
	ctx context.Context,
	app *pb.Application,
) (*empty.Empty, error) {
	if _, ok := s.containerApps[app.Id]; ok {
		s.containerApps[app.Id] = app
		return &empty.Empty{}, nil
	}

	if _, ok := s.vmApps[app.Id]; ok {
		s.vmApps[app.Id] = app
		return &empty.Empty{}, nil
	}

	return nil, status.Errorf(
		codes.NotFound, "Application %s not found", app.Id)
}

func (s *applicationServer) Remove(
	ctx context.Context,
	id *pb.ApplicationID,
) (*empty.Empty, error) {
	var policyExists bool
	if _, policyExists = s.policyServer.policies[id.Id]; policyExists {
		delete(s.policyServer.policies, id.Id)
		policyExists = true
	}

	if _, ok := s.containerApps[id.Id]; ok {
		delete(s.containerApps, id.Id)
		return &empty.Empty{}, nil
	}

	if _, ok := s.vmApps[id.Id]; ok {
		delete(s.vmApps, id.Id)
		return &empty.Empty{}, nil
	}

	if policyExists {
		return nil, status.Errorf(codes.DataLoss,
			"Application %s not found but had a policy!", id.Id)
	}

	return nil, status.Errorf(codes.NotFound, "Application %s not found", id.Id)
}

func (s *applicationServer) Start(
	ctx context.Context,
	cmd *pb.LifecycleCommand,
) (*empty.Empty, error) {
	app := s.find(cmd.Id)

	if app != nil {
		if app.Status != pb.LifecycleStatus_STOPPED {
			return nil, status.Errorf(
				codes.FailedPrecondition, "Application %s not stopped", cmd.Id)
		}

		app.Status = pb.LifecycleStatus_RUNNING
		return &empty.Empty{}, nil
	}

	return nil, status.Errorf(
		codes.NotFound, "Application %s not found", cmd.Id)
}

func (s *applicationServer) Stop(
	ctx context.Context,
	cmd *pb.LifecycleCommand,
) (*empty.Empty, error) {
	app := s.find(cmd.Id)

	if app != nil {
		if app.Status != pb.LifecycleStatus_RUNNING {
			return nil, status.Errorf(
				codes.FailedPrecondition, "Application %s not running", cmd.Id)
		}

		app.Status = pb.LifecycleStatus_STOPPED
		return &empty.Empty{}, nil
	}

	return nil, status.Errorf(
		codes.NotFound, "Application %s not found", cmd.Id)
}

func (s *applicationServer) Restart(
	ctx context.Context,
	cmd *pb.LifecycleCommand,
) (*empty.Empty, error) {
	app := s.find(cmd.Id)

	if app != nil {
		if app.Status != pb.LifecycleStatus_RUNNING {
			return nil, status.Errorf(
				codes.FailedPrecondition, "Application %s not running", cmd.Id)
		}

		return &empty.Empty{}, nil
	}

	return nil, status.Errorf(
		codes.NotFound, "Application %s not found", cmd.Id)
}

func (s *applicationServer) find(id string) *pb.Application {
	if containerApp, ok := s.containerApps[id]; ok {
		return containerApp
	}

	if vmApp, ok := s.vmApps[id]; ok {
		return vmApp
	}

	return nil
}
