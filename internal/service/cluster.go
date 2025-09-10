package service

import "star-dim/internal/models"

type ClusterService struct{}

func NewClusterServer() *ClusterService {
	return &ClusterService{}
}

func (s *ClusterService) GetClusters() []*models.Cluster {
	// 从数据库读取所有cluster信息
	// test
	clusters := []*models.Cluster{}
	clusters = append(clusters, &models.Cluster{
		Name: "hpc1",
		LoginNodes: []*models.LoginNode{
			{
				Name: "ln1",
				Host: "1.94.239.51",
				Port: "22",
			},
		},
	})
	return clusters
}

func (s *ClusterService) GetCluster(name string) *models.Cluster {
	return nil
}
