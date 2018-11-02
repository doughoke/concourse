package db_test

import (
	"time"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/creds"
	"github.com/concourse/concourse/atc/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourceConfig", func() {
	Describe("AcquireResourceConfigCheckingLockWithIntervalCheck", func() {
		var (
			someResource               db.Resource
			resourceConfigCheckSession db.ResourceConfigCheckSession
			resourceConfig             db.ResourceConfig
		)

		ownerExpiries := db.ContainerOwnerExpiries{
			GraceTime: 1 * time.Minute,
			Min:       5 * time.Minute,
			Max:       5 * time.Minute,
		}

		BeforeEach(func() {
			var err error
			var found bool

			resourceConfigFactory = db.NewResourceConfigFactory(dbConn, lockFactory)
			someResource, found, err = defaultPipeline.Resource("some-resource")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			pipelineResourceTypes, err := defaultPipeline.ResourceTypes()
			Expect(err).ToNot(HaveOccurred())

			resourceConfigCheckSession, err = resourceConfigCheckSessionFactory.FindOrCreateResourceConfigCheckSession(
				logger,
				someResource.Type(),
				someResource.Source(),
				creds.NewVersionedResourceTypes(template.StaticVariables{}, pipelineResourceTypes.Deserialize()),
				ownerExpiries,
			)
			Expect(err).ToNot(HaveOccurred())

			resourceConfig = resourceConfigCheckSession.ResourceConfig()
		})

		Context("when there has been a check recently", func() {
			Context("when acquiring immediately", func() {
				It("gets the lock", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					lock, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when not acquiring immediately", func() {
				It("does not get the lock", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					_, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeFalse())
				})
			})
		})

		Context("when there has not been a check recently", func() {
			Context("when acquiring immediately", func() {
				It("gets and keeps the lock and stops others from periodically getting it", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					Consistently(func() bool {
						_, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
						Expect(err).ToNot(HaveOccurred())

						return acquired
					}, 1500*time.Millisecond, 100*time.Millisecond).Should(BeFalse())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(time.Second)

					lock, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())
				})

				It("gets and keeps the lock and stops others from immediately getting it", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					Consistently(func() bool {
						_, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
						Expect(err).ToNot(HaveOccurred())

						return acquired
					}, 1500*time.Millisecond, 100*time.Millisecond).Should(BeFalse())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(time.Second)

					lock, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when not acquiring immediately", func() {
				It("gets and keeps the lock and stops others from periodically getting it", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					Consistently(func() bool {
						_, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
						Expect(err).ToNot(HaveOccurred())

						return acquired
					}, 1500*time.Millisecond, 100*time.Millisecond).Should(BeFalse())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(time.Second)

					lock, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())
				})

				It("gets and keeps the lock and stops others from immediately getting it", func() {
					lock, acquired, err := resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)

					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					Consistently(func() bool {
						_, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, true)
						Expect(err).ToNot(HaveOccurred())

						return acquired
					}, 1500*time.Millisecond, 100*time.Millisecond).Should(BeFalse())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(time.Second)

					lock, acquired, err = resourceConfig.AcquireResourceConfigCheckingLockWithIntervalCheck(logger, 1*time.Second, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(acquired).To(BeTrue())

					err = lock.Release()
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("SaveVersions", func() {
		var (
			originalVersionSlice []atc.Version
			resourceConfig       db.ResourceConfig
		)

		BeforeEach(func() {
			setupTx, err := dbConn.Begin()
			Expect(err).ToNot(HaveOccurred())

			brt := db.BaseResourceType{
				Name: "some-type",
			}
			_, err = brt.FindOrCreate(setupTx)
			Expect(err).NotTo(HaveOccurred())
			Expect(setupTx.Commit()).To(Succeed())

			resourceConfigFactory := db.NewResourceConfigFactory(dbConn, lockFactory)
			resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
			Expect(err).ToNot(HaveOccurred())

			originalVersionSlice = []atc.Version{
				{"ref": "v1"},
				{"ref": "v3"},
			}
		})

		// XXX: Can make test more resilient if there is a method that gives all versions by descending check order
		It("ensures versioned resources have the correct check_order", func() {
			err := resourceConfig.SaveVersions(originalVersionSlice)
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpace(atc.Space("space"))
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpaceLatestVersion(atc.Space("space"), atc.Version{"ref": "v3"})
			Expect(err).ToNot(HaveOccurred())

			latestVR, found, err := resourceConfig.LatestVersions()
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(latestVR[0].Version()).To(Equal(db.Version{"ref": "v3"}))
			Expect(latestVR[0].CheckOrder()).To(Equal(2))

			pretendCheckResults := []atc.Version{
				{"ref": "v2"},
				{"ref": "v3"},
			}

			err = resourceConfig.SaveVersions(pretendCheckResults)
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpaceLatestVersion(atc.Space("space"), atc.Version{"ref": "v3"})
			Expect(err).ToNot(HaveOccurred())

			latestVR, found, err = resourceConfig.LatestVersions()
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(latestVR[0].Version()).To(Equal(db.Version{"ref": "v3"}))
			Expect(latestVR[0].CheckOrder()).To(Equal(4))
		})

		Context("when the versions already exists", func() {
			var newVersionSlice []atc.Version

			BeforeEach(func() {
				newVersionSlice = []atc.Version{
					{"ref": "v1"},
					{"ref": "v3"},
				}
			})

			It("does not change the check order", func() {
				err := resourceConfig.SaveVersions(newVersionSlice)
				Expect(err).ToNot(HaveOccurred())

				err = resourceConfig.SaveSpace(atc.Space("space"))
				Expect(err).ToNot(HaveOccurred())

				err = resourceConfig.SaveSpaceLatestVersion(atc.Space("space"), atc.Version{"ref": "v3"})
				Expect(err).ToNot(HaveOccurred())

				latestVR, found, err := resourceConfig.LatestVersions()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())

				Expect(latestVR[0].Version()).To(Equal(db.Version{"ref": "v3"}))
				Expect(latestVR[0].CheckOrder()).To(Equal(2))
			})
		})
	})

	Describe("LatestVersions", func() {
		var (
			originalVersionSlice []atc.Version
			resourceConfig       db.ResourceConfig
			latestCV             []db.ResourceConfigVersion
			found                bool
		)

		Context("when the resource config exists", func() {
			BeforeEach(func() {
				setupTx, err := dbConn.Begin()
				Expect(err).ToNot(HaveOccurred())

				brt := db.BaseResourceType{
					Name: "some-type",
				}
				_, err = brt.FindOrCreate(setupTx)
				Expect(err).NotTo(HaveOccurred())
				Expect(setupTx.Commit()).To(Succeed())

				resourceConfigFactory := db.NewResourceConfigFactory(dbConn, lockFactory)
				resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
				Expect(err).ToNot(HaveOccurred())

				originalVersionSlice = []atc.Version{
					{"ref": "v1"},
					{"ref": "v3"},
				}

				err = resourceConfig.SaveVersions(originalVersionSlice)
				Expect(err).ToNot(HaveOccurred())

				err = resourceConfig.SaveSpace(atc.Space("space"))
				Expect(err).ToNot(HaveOccurred())

				err = resourceConfig.SaveSpaceLatestVersion(atc.Space("space"), atc.Version{"ref": "v3"})
				Expect(err).ToNot(HaveOccurred())

				latestCV, found, err = resourceConfig.LatestVersions()
				Expect(err).ToNot(HaveOccurred())
			})

			It("gets latest version of resource", func() {
				Expect(found).To(BeTrue())

				Expect(latestCV[0].Version()).To(Equal(db.Version{"ref": "v3"}))
				Expect(latestCV[0].CheckOrder()).To(Equal(2))
			})
		})
	})

	Describe("FindVersion", func() {
		var (
			originalVersionSlice []atc.Version
			resourceConfig       db.ResourceConfig
			latestCV             db.ResourceConfigVersion
			found                bool
		)

		BeforeEach(func() {
			setupTx, err := dbConn.Begin()
			Expect(err).ToNot(HaveOccurred())

			brt := db.BaseResourceType{
				Name: "some-type",
			}
			_, err = brt.FindOrCreate(setupTx)
			Expect(err).NotTo(HaveOccurred())
			Expect(setupTx.Commit()).To(Succeed())

			resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
			Expect(err).ToNot(HaveOccurred())

			originalVersionSlice = []atc.Version{
				{"ref": "v1"},
				{"ref": "v3"},
			}

			err = resourceConfig.SaveVersions(originalVersionSlice)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the version exists", func() {
			BeforeEach(func() {
				var err error
				latestCV, found, err = resourceConfig.FindVersion(atc.Version{"ref": "v1"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("gets the version of resource", func() {
				Expect(found).To(BeTrue())

				Expect(latestCV.ResourceConfig().ID()).To(Equal(resourceConfig.ID()))
				Expect(latestCV.Version()).To(Equal(db.Version{"ref": "v1"}))
				Expect(latestCV.CheckOrder()).To(Equal(1))
			})
		})

		Context("when the version does not exist", func() {
			BeforeEach(func() {
				var err error
				latestCV, found, err = resourceConfig.FindVersion(atc.Version{"ref": "v2"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not get the version of resource", func() {
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("SaveDefaultSpace", func() {
		var (
			resourceConfig  db.ResourceConfig
			defaultSpace    string
			defaultSpaceErr error
		)

		BeforeEach(func() {
			setupTx, err := dbConn.Begin()
			Expect(err).ToNot(HaveOccurred())

			brt := db.BaseResourceType{
				Name: "some-type",
			}
			_, err = brt.FindOrCreate(setupTx)
			Expect(err).NotTo(HaveOccurred())
			Expect(setupTx.Commit()).To(Succeed())

			resourceConfigFactory := db.NewResourceConfigFactory(dbConn, lockFactory)
			resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			defaultSpaceErr = resourceConfig.SaveDefaultSpace(atc.Space(defaultSpace))
		})

		Context("when the space exists", func() {
			BeforeEach(func() {
				err := resourceConfig.SaveSpace(atc.Space("space"))
				Expect(err).ToNot(HaveOccurred())

				defaultSpace = "space"
			})

			It("saves the default space", func() {
				Expect(defaultSpaceErr).ToNot(HaveOccurred())

				resourceConfig, err := resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
				Expect(err).ToNot(HaveOccurred())
				Expect(resourceConfig.DefaultSpace()).To(Equal(atc.Space("space")))
			})
		})
	})

	Describe("SaveVersion/SaveSpace", func() {
		var (
			resourceConfig db.ResourceConfig
			spaceVersion   atc.SpaceVersion
		)

		BeforeEach(func() {
			setupTx, err := dbConn.Begin()
			Expect(err).ToNot(HaveOccurred())

			brt := db.BaseResourceType{
				Name: "some-type",
			}
			_, err = brt.FindOrCreate(setupTx)
			Expect(err).NotTo(HaveOccurred())
			Expect(setupTx.Commit()).To(Succeed())

			resourceConfigFactory := db.NewResourceConfigFactory(dbConn, lockFactory)
			resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
			Expect(err).ToNot(HaveOccurred())

			spaceVersion = atc.SpaceVersion{
				Space:   "space",
				Version: atc.Version{"some": "version"},
				Metadata: atc.Metadata{
					atc.MetadataField{
						Name:  "some",
						Value: "metadata",
					},
				},
			}
		})

		It("saves the version if the space exists", func() {
			err := resourceConfig.SaveSpace(atc.Space("space"))
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveVersion(spaceVersion)
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpaceLatestVersion(atc.Space("space"), atc.Version{"some": "version"})
			Expect(err).ToNot(HaveOccurred())

			latestVR, found, err := resourceConfig.LatestVersions()
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(latestVR[0].Version()).To(Equal(db.Version{"some": "version"}))
			Expect(latestVR[0].CheckOrder()).To(Equal(1))
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				spaceVersion = atc.SpaceVersion{
					Space:   "unknown-space",
					Version: atc.Version{"some": "version"},
					Metadata: atc.Metadata{
						atc.MetadataField{
							Name:  "some",
							Value: "metadata",
						},
					},
				}
			})

			It("does not save the version", func() {
				err := resourceConfig.SaveVersion(spaceVersion)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("SaveSpaceLatestVersion/LatestVersions", func() {
		var (
			resourceConfig db.ResourceConfig
			spaceVersion   atc.SpaceVersion
			spaceVersion2  atc.SpaceVersion
		)

		BeforeEach(func() {
			setupTx, err := dbConn.Begin()
			Expect(err).ToNot(HaveOccurred())

			brt := db.BaseResourceType{
				Name: "some-type",
			}
			_, err = brt.FindOrCreate(setupTx)
			Expect(err).NotTo(HaveOccurred())
			Expect(setupTx.Commit()).To(Succeed())

			resourceConfigFactory := db.NewResourceConfigFactory(dbConn, lockFactory)
			resourceConfig, err = resourceConfigFactory.FindOrCreateResourceConfig(logger, "some-type", atc.Source{"source-config": "some-value"}, creds.VersionedResourceTypes{})
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpace(atc.Space("space"))
			Expect(err).ToNot(HaveOccurred())

			otherSpaceVersion := atc.SpaceVersion{
				Space:   "space",
				Version: atc.Version{"some": "other-version"},
				Metadata: atc.Metadata{
					atc.MetadataField{
						Name:  "some",
						Value: "metadata",
					},
				},
			}

			err = resourceConfig.SaveVersion(otherSpaceVersion)
			Expect(err).ToNot(HaveOccurred())

			spaceVersion = atc.SpaceVersion{
				Space:   "space",
				Version: atc.Version{"some": "version"},
				Metadata: atc.Metadata{
					atc.MetadataField{
						Name:  "some",
						Value: "metadata",
					},
				},
			}

			err = resourceConfig.SaveVersion(spaceVersion)
			Expect(err).ToNot(HaveOccurred())

			err = resourceConfig.SaveSpace(atc.Space("space2"))
			Expect(err).ToNot(HaveOccurred())

			spaceVersion2 = atc.SpaceVersion{
				Space:   "space2",
				Version: atc.Version{"some": "version2"},
				Metadata: atc.Metadata{
					atc.MetadataField{
						Name:  "some",
						Value: "metadata",
					},
				},
			}

			err = resourceConfig.SaveVersion(spaceVersion2)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the version exists", func() {
			BeforeEach(func() {
				err := resourceConfig.SaveSpaceLatestVersion(spaceVersion.Space, spaceVersion.Version)
				Expect(err).ToNot(HaveOccurred())

				err = resourceConfig.SaveSpaceLatestVersion(spaceVersion2.Space, spaceVersion2.Version)
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the version into the space", func() {
				latestVersions, found, err := resourceConfig.LatestVersions()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(latestVersions).To(HaveLen(2))
				Expect(latestVersions[0].Version()).To(Equal(db.Version(spaceVersion.Version)))
				Expect(latestVersions[1].Version()).To(Equal(db.Version(spaceVersion2.Version)))
			})
		})
	})
})
