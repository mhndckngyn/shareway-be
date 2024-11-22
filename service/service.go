package service

import (
	"shareway/infra/bucket"
	"shareway/infra/fpt"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/util"
	"shareway/util/sanctum"
	"shareway/util/token"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContainer struct {
	OTPService          IOTPService
	UserService         IUsersService
	MapService          IMapService
	VehicleService      IVehicleService
	RideService         IRideService
	NotificationService INotificationService
	ChatService         IChatService
}

type ServiceFactory struct {
	repos     *repository.RepositoryContainer
	cfg       util.Config
	fptReader *fpt.FPTReader
	encryptor util.IEncryptor
	maker     *token.PasetoMaker
	redis     *redis.Client
	hub       *ws.Hub
	// rabbitmq   *rabbitmq.RabbitMQ
	asynq        *task.AsyncClient
	cloudinary   *bucket.CloudinaryService
	sanctumToken *sanctum.SanctumToken
}

func NewServiceFactory(db *gorm.DB, cfg util.Config, token *token.PasetoMaker, redisClient *redis.Client, hub *ws.Hub, asynq *task.AsyncClient, cloudinary *bucket.CloudinaryService) *ServiceFactory {
	repoFactory := repository.NewRepositoryFactory(db, redisClient, cfg)
	repos := repoFactory.CreateRepositories()

	// Initialize FPT reader
	fptReader := fpt.NewFPTReader(cfg)
	// Initialize encryptor
	encryptor := util.NewEncryptor(cfg)

	// Initialize the token
	cryptoSanctum := sanctum.NewCryptoSanctum()
	tokenSanctum := sanctum.NewTokenSanctum(cryptoSanctum)
	sanctumToken := sanctum.NewSanctumToken(tokenSanctum, db)

	return &ServiceFactory{
		repos:        repos,
		cfg:          cfg,
		fptReader:    fptReader,
		encryptor:    encryptor,
		maker:        token,
		redis:        redisClient,
		hub:          hub,
		cloudinary:   cloudinary,
		asynq:        asynq,
		sanctumToken: sanctumToken,
	}
}

func (f *ServiceFactory) CreateServices() *ServiceContainer {
	return &ServiceContainer{
		OTPService:          f.createOTPService(),
		UserService:         f.createUserService(),
		MapService:          f.createMapsService(),
		VehicleService:      f.createVehicleService(),
		RideService:         f.createRideService(),
		NotificationService: f.createNotificationService(),
		ChatService:         f.createChatService(),
	}
}

func (f *ServiceFactory) createOTPService() IOTPService {
	return NewOTPService(f.cfg, f.repos.OTPRepository)
}

func (f *ServiceFactory) createUserService() IUsersService {
	return NewUsersService(f.repos.AuthRepository, f.encryptor, f.fptReader, f.maker, f.cfg)
}

func (f *ServiceFactory) createMapsService() IMapService {
	return NewMapService(f.repos.MapsRepository, f.cfg, f.redis)
}

func (f *ServiceFactory) createVehicleService() IVehicleService {
	return NewVehicleService(f.repos.VehicleRepository, f.cfg)
}

func (f *ServiceFactory) createRideService() IRideService {
	return NewRideService(f.repos.RideRepository, f.hub, f.cfg)
}

func (f *ServiceFactory) createNotificationService() INotificationService {
	return NewNotificationService(f.repos.NotificationRepository, f.cfg, f.asynq)
}

func (f *ServiceFactory) createChatService() IChatService {
	return NewChatService(f.repos.ChatRepository, f.hub, f.cfg, f.cloudinary)
}
