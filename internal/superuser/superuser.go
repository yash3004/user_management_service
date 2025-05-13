package superuser

import (
	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/schemas"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

// SuperUserConfig holds configuration for the super user
type SuperUserConfig struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// DefaultSuperUserConfig returns default configuration for super user
func DefaultSuperUserConfig() SuperUserConfig {
	return SuperUserConfig{
		Email:     "admin@example.com",
		Password:  "superuser123", // This should be changed in production
		FirstName: "Super",
		LastName:  "User",
	}
}

// EnsureSuperUser creates a super user if it doesn't exist
func EnsureSuperUser(db *gorm.DB, config SuperUserConfig) error {
	// Check if super user role exists
	var superRole schemas.Role
	if err := db.Where("name = ?", "SuperAdmin").First(&superRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create super role
			superRole = schemas.Role{
				ID:          uuid.New(),
				Name:        "SuperAdmin",
				Description: "Super administrator with full access to all resources",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := db.Create(&superRole).Error; err != nil {
				klog.Errorf("Failed to create super admin role: %v", err)
				return err
			}
			klog.Info("Created SuperAdmin role")
		} else {
			klog.Errorf("Error checking for SuperAdmin role: %v", err)
			return err
		}
	}

	// Create policies for the super role
	policies := []schemas.Policy{
		{
			ID:          uuid.New(),
			Name:        "AllPoliciesAccess",
			Description: "Full access to manage policies",
			Resource:    "policies",
			Action:      "*",
			Effect:      "allow",
			RolesId:     superRole.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "AllRolesAccess",
			Description: "Full access to manage roles",
			Resource:    "roles",
			Action:      "*",
			Effect:      "allow",
			RolesId:     superRole.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "AllProjectsAccess",
			Description: "Full access to manage projects",
			Resource:    "projects",
			Action:      "*",
			Effect:      "allow",
			RolesId:     superRole.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Create policies if they don't exist
	for _, policy := range policies {
		var existingPolicy schemas.Policy
		if err := db.Where("name = ?", policy.Name).First(&existingPolicy).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&policy).Error; err != nil {
					klog.Errorf("Failed to create policy %s: %v", policy.Name, err)
					return err
				}
				klog.Infof("Created policy: %s", policy.Name)
			} else {
				klog.Errorf("Error checking for policy %s: %v", policy.Name, err)
				return err
			}
		}
	}

	// Check if default project exists
	var defaultProject schemas.Project
	if err := db.Where("unique_id = ?", "default").First(&defaultProject).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default project
			defaultProject = schemas.Project{
				ID:          uuid.New(),
				Name:        "Default Project",
				Description: "Default project for system administration",
				UniqueID:    "default",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := db.Create(&defaultProject).Error; err != nil {
				klog.Errorf("Failed to create default project: %v", err)
				return err
			}
			klog.Info("Created default project")
		} else {
			klog.Errorf("Error checking for default project: %v", err)
			return err
		}
	}

	// Check if super user exists
	var superUser schemas.User
	if err := db.Where("email = ?", config.Email).First(&superUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
			if err != nil {
				klog.Errorf("Failed to hash password: %v", err)
				return err
			}

			// Create super user
			superUser = schemas.User{
				ID:        uuid.New(),
				Email:     config.Email,
				Password:  string(hashedPassword),
				FirstName: config.FirstName,
				LastName:  config.LastName,
				Active:    true,
				RoleId:    superRole.ID,
				ProjectId: defaultProject.ID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.Create(&superUser).Error; err != nil {
				klog.Errorf("Failed to create super user: %v", err)
				return err
			}
			klog.Infof("Created super user with email: %s", config.Email)
		} else {
			klog.Errorf("Error checking for super user: %v", err)
			return err
		}
	} else {
		klog.Info("Super user already exists")
	}

	return nil
}