package service

import (
	"gpanel/dto"
	"gpanel/models"
	"gpanel/repo"
	"gpanel/utils/encrypt"
	"time"
)

type HostService struct {
	hostRepo repo.IHostRepo
}

type IHostService interface {
	// HostGroup operations
	CreateGroup(req dto.HostGroupOperate) (*models.HostGroup, error)
	UpdateGroup(req dto.HostGroupOperate) error
	DeleteGroup(id uint) error
	GetGroupByID(id uint) (*dto.HostGroupInfo, error)
	ListGroups(req dto.HostGroupSearch) ([]dto.HostGroupInfo, int64, error)
	GetHostTree() ([]dto.HostTreeNode, error)

	// Host operations
	CreateHost(req dto.HostOperate) (*dto.HostInfo, error)
	UpdateHost(req dto.HostOperate) error
	DeleteHost(id uint) error
	GetHostByID(id uint) (*dto.HostInfo, error)
	GetHostForConnection(id uint) (*models.Host, error)
	ListHosts(req dto.HostSearch) ([]dto.HostInfo, int64, error)
	MoveHosts(req dto.HostMove) error
	ExportHosts(encrypted bool) ([]dto.HostOperate, error)
	ImportHosts(hosts []dto.HostOperate) (int, int, error)
}

func NewHostService() IHostService {
	return &HostService{
		hostRepo: repo.NewHostRepo(),
	}
}

// HostGroup operations

func (s *HostService) CreateGroup(req dto.HostGroupOperate) (*models.HostGroup, error) {
	group := &models.HostGroup{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.hostRepo.CreateGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *HostService) UpdateGroup(req dto.HostGroupOperate) error {
	group, err := s.hostRepo.GetGroupByID(req.ID)
	if err != nil {
		return err
	}
	group.Name = req.Name
	group.Description = req.Description
	return s.hostRepo.UpdateGroup(group)
}

func (s *HostService) DeleteGroup(id uint) error {
	return s.hostRepo.DeleteGroup(id)
}

func (s *HostService) GetGroupByID(id uint) (*dto.HostGroupInfo, error) {
	group, err := s.hostRepo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}
	count, _ := s.hostRepo.GetGroupHostCount(id)
	return &dto.HostGroupInfo{
		ID:          group.ID,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
		Name:        group.Name,
		Description: group.Description,
		HostCount:   int(count),
	}, nil
}

func (s *HostService) ListGroups(req dto.HostGroupSearch) ([]dto.HostGroupInfo, int64, error) {
	groups, total, err := s.hostRepo.SearchGroups(req.Page, req.PageSize, req.Info)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.HostGroupInfo
	for _, group := range groups {
		count, _ := s.hostRepo.GetGroupHostCount(group.ID)
		result = append(result, dto.HostGroupInfo{
			ID:          group.ID,
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
			Name:        group.Name,
			Description: group.Description,
			HostCount:   int(count),
		})
	}
	return result, total, nil
}

func (s *HostService) GetHostTree() ([]dto.HostTreeNode, error) {
	groups, err := s.hostRepo.ListGroups()
	if err != nil {
		return nil, err
	}

	var tree []dto.HostTreeNode
	for _, group := range groups {
		hosts, _ := s.hostRepo.GetHostsByGroupID(group.ID)
		var children []dto.HostTreeNode
		for _, host := range hosts {
			children = append(children, dto.HostTreeNode{
				ID:               host.ID,
				Name:             host.Name,
				Type:             "host",
				GroupID:          host.GroupID,
				Addr:             host.Addr,
				Port:             host.Port,
				User:             host.User,
				AuthMode:         host.AuthMode,
				RememberPassword: host.RememberPassword,
				Description:      host.Description,
			})
		}
		tree = append(tree, dto.HostTreeNode{
			ID:       group.ID,
			Name:     group.Name,
			Type:     "group",
			Children: children,
		})
	}
	return tree, nil
}

// Host operations

func (s *HostService) CreateHost(req dto.HostOperate) (*dto.HostInfo, error) {
	host := &models.Host{
		GroupID:          req.GroupID,
		Name:             req.Name,
		Addr:             req.Addr,
		Port:             req.Port,
		User:             req.User,
		AuthMode:         req.AuthMode,
		RememberPassword: req.RememberPassword,
		Description:      req.Description,
	}

	// 加密敏感信息
	if req.AuthMode == "password" && req.Password != "" {
		encrypted, err := encrypt.StringEncrypt(req.Password)
		if err != nil {
			return nil, err
		}
		host.Password = encrypted
	} else if req.AuthMode == "key" {
		if req.PrivateKey != "" {
			encrypted, err := encrypt.StringEncrypt(req.PrivateKey)
			if err != nil {
				return nil, err
			}
			host.PrivateKey = encrypted
		}
		if req.PassPhrase != "" {
			encrypted, err := encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				return nil, err
			}
			host.PassPhrase = encrypted
		}
	}

	if err := s.hostRepo.CreateHost(host); err != nil {
		return nil, err
	}

	return s.convertToHostInfo(host), nil
}

func (s *HostService) UpdateHost(req dto.HostOperate) error {
	host, err := s.hostRepo.GetHostByID(req.ID)
	if err != nil {
		return err
	}

	host.GroupID = req.GroupID
	host.Name = req.Name
	host.Addr = req.Addr
	host.Port = req.Port
	host.User = req.User
	host.AuthMode = req.AuthMode
	host.RememberPassword = req.RememberPassword
	host.Description = req.Description

	// 加密敏感信息
	if req.AuthMode == "password" && req.Password != "" {
		encrypted, err := encrypt.StringEncrypt(req.Password)
		if err != nil {
			return err
		}
		host.Password = encrypted
		host.PrivateKey = ""
		host.PassPhrase = ""
	} else if req.AuthMode == "key" {
		if req.PrivateKey != "" {
			encrypted, err := encrypt.StringEncrypt(req.PrivateKey)
			if err != nil {
				return err
			}
			host.PrivateKey = encrypted
		}
		if req.PassPhrase != "" {
			encrypted, err := encrypt.StringEncrypt(req.PassPhrase)
			if err != nil {
				return err
			}
			host.PassPhrase = encrypted
		}
		host.Password = ""
	}

	return s.hostRepo.UpdateHost(host)
}

func (s *HostService) DeleteHost(id uint) error {
	return s.hostRepo.DeleteHost(id)
}

func (s *HostService) GetHostByID(id uint) (*dto.HostInfo, error) {
	host, err := s.hostRepo.GetHostByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToHostInfo(host), nil
}

func (s *HostService) GetHostForConnection(id uint) (*models.Host, error) {
	host, err := s.hostRepo.GetHostByID(id)
	if err != nil {
		return nil, err
	}

	// 解密敏感信息
	if host.AuthMode == "password" && host.Password != "" {
		decrypted, err := encrypt.StringDecrypt(host.Password)
		if err != nil {
			return nil, err
		}
		host.Password = decrypted
	} else if host.AuthMode == "key" {
		if host.PrivateKey != "" {
			decrypted, err := encrypt.StringDecrypt(host.PrivateKey)
			if err != nil {
				return nil, err
			}
			host.PrivateKey = decrypted
		}
		if host.PassPhrase != "" {
			decrypted, err := encrypt.StringDecrypt(host.PassPhrase)
			if err != nil {
				return nil, err
			}
			host.PassPhrase = decrypted
		}
	}

	return host, nil
}

func (s *HostService) ListHosts(req dto.HostSearch) ([]dto.HostInfo, int64, error) {
	hosts, total, err := s.hostRepo.SearchHosts(req.Page, req.PageSize, int(req.GroupID), req.Info)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.HostInfo
	for _, host := range hosts {
		result = append(result, *s.convertToHostInfo(&host))
	}
	return result, total, nil
}

func (s *HostService) MoveHosts(req dto.HostMove) error {
	return s.hostRepo.MoveHostsToGroup(req.HostIDs, req.GroupID)
}

// ExportHosts 导出主机列表
func (s *HostService) ExportHosts(encrypted bool) ([]dto.HostOperate, error) {
	hosts, err := s.hostRepo.ListHosts()
	if err != nil {
		return nil, err
	}

	var result []dto.HostOperate
	for _, host := range hosts {
		exportHost := dto.HostOperate{
			ID:               host.ID,
			GroupID:          host.GroupID,
			Name:             host.Name,
			Addr:             host.Addr,
			Port:             host.Port,
			User:             host.User,
			AuthMode:         host.AuthMode,
			RememberPassword: host.RememberPassword,
			Description:      host.Description,
		}

		// 处理敏感信息
		if host.AuthMode == "password" && host.Password != "" {
			decrypted, err := encrypt.StringDecrypt(host.Password)
			if err == nil {
				if encrypted {
					// 加密导出：使用加密后的值
					exportHost.Password = host.Password
				} else {
					// 明文导出：使用解密后的值
					exportHost.Password = decrypted
				}
			}
		} else if host.AuthMode == "key" {
			if host.PrivateKey != "" {
				decrypted, err := encrypt.StringDecrypt(host.PrivateKey)
				if err == nil {
					if encrypted {
						exportHost.PrivateKey = host.PrivateKey
					} else {
						exportHost.PrivateKey = decrypted
					}
				}
			}
			if host.PassPhrase != "" {
				decrypted, err := encrypt.StringDecrypt(host.PassPhrase)
				if err == nil {
					if encrypted {
						exportHost.PassPhrase = host.PassPhrase
					} else {
						exportHost.PassPhrase = decrypted
					}
				}
			}
		}

		result = append(result, exportHost)
	}
	return result, nil
}

// ImportHosts 导入主机列表
func (s *HostService) ImportHosts(hosts []dto.HostOperate) (int, int, error) {
	var successCount, failCount int
	for _, host := range hosts {
		// 检查是否已存在相同名称的主机
		existing, err := s.hostRepo.ListHosts()
		if err == nil {
			for _, e := range existing {
				if e.Name == host.Name && e.Addr == host.Addr {
					failCount++
					continue
				}
			}
		}

		// 处理敏感信息：尝试解密，如果失败则认为是明文，需要加密存储
		importHost := host
		if host.AuthMode == "password" && host.Password != "" {
			// 尝试解密，判断是否已加密
			_, err := encrypt.StringDecrypt(host.Password)
			if err != nil {
				// 解密失败，说明是明文，需要加密存储
				encrypted, err := encrypt.StringEncrypt(host.Password)
				if err == nil {
					importHost.Password = encrypted
				}
			} else {
				// 解密成功，说明已加密，直接使用原值（已经是加密状态）
				importHost.Password = host.Password
			}
		} else if host.AuthMode == "key" {
			if host.PrivateKey != "" {
				_, err := encrypt.StringDecrypt(host.PrivateKey)
				if err != nil {
					// 明文，需要加密
					encrypted, err := encrypt.StringEncrypt(host.PrivateKey)
					if err == nil {
						importHost.PrivateKey = encrypted
					}
				} else {
					// 已加密
					importHost.PrivateKey = host.PrivateKey
				}
			}
			if host.PassPhrase != "" {
				_, err := encrypt.StringDecrypt(host.PassPhrase)
				if err != nil {
					// 明文，需要加密
					encrypted, err := encrypt.StringEncrypt(host.PassPhrase)
					if err == nil {
						importHost.PassPhrase = encrypted
					}
				} else {
					// 已加密
					importHost.PassPhrase = host.PassPhrase
				}
			}
		}

		// 创建新主机（不再通过 CreateHost，因为它会再次加密）
		hostModel := &models.Host{
			GroupID:          importHost.GroupID,
			Name:             importHost.Name,
			Addr:             importHost.Addr,
			Port:             importHost.Port,
			User:             importHost.User,
			AuthMode:         importHost.AuthMode,
			Password:         importHost.Password,
			PrivateKey:       importHost.PrivateKey,
			PassPhrase:       importHost.PassPhrase,
			RememberPassword: importHost.RememberPassword,
			Description:      importHost.Description,
		}

		if err := s.hostRepo.CreateHost(hostModel); err != nil {
			failCount++
		} else {
			successCount++
		}
	}
	return successCount, failCount, nil
}

func (s *HostService) convertToHostInfo(host *models.Host) *dto.HostInfo {
	groupName := ""
	if host.HostGroup != nil {
		groupName = host.HostGroup.Name
	}

	return &dto.HostInfo{
		ID:               host.ID,
		CreatedAt:        host.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        host.UpdatedAt.Format(time.RFC3339),
		GroupID:          host.GroupID,
		GroupName:        groupName,
		Name:             host.Name,
		Addr:             host.Addr,
		Port:             host.Port,
		User:             host.User,
		AuthMode:         host.AuthMode,
		RememberPassword: host.RememberPassword,
		Description:      host.Description,
	}
}
