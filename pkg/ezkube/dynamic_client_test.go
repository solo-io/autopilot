package ezkube_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/autopilot/pkg/ezkube"
	"github.com/solo-io/autopilot/test"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("SimpleClient", func() {
	var (
		mgr    manager.Manager
		cancel context.CancelFunc
	)
	BeforeEach(func() {
		mgr, cancel = test.MustManager()
	})
	AfterEach(func() {
		cancel()
	})
	It("ensures resources to the cluster", func() {
		r := rand.String(4)
		parentName := "parent-" + r
		childName := "child-" + r

		ns := "default"
		parent := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      parentName,
			},
			Data: map[string]string{"some": "data"},
		}

		child := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      childName,
			},
			Data: map[string]string{"smore": "data"},
		}

		client := NewClient(mgr)
		err := client.Create(context.TODO(), parent)
		Expect(err).NotTo(HaveOccurred())

		err = client.Ensure(context.TODO(), parent, child)
		Expect(err).NotTo(HaveOccurred())

		actualParent := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      parentName,
			},
		}
		err = client.Get(context.TODO(), actualParent)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualParent.Data).To(Equal(parent.Data))

		actualChild := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      childName,
			},
		}
		Eventually(func() error {
			err = client.Get(context.TODO(), actualChild)
			return err
		}).Should(Not(HaveOccurred()))
		Expect(actualChild.Data).To(Equal(child.Data))
		Expect(actualChild.OwnerReferences).To(HaveLen(1))
		t := true
		Expect(actualChild.OwnerReferences[0]).To(Equal(metav1.OwnerReference{
			APIVersion:         "v1",
			Kind:               "ConfigMap",
			Name:               actualParent.Name,
			UID:                actualParent.UID,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}))

		err = client.Delete(context.TODO(), actualParent)
		Expect(err).NotTo(HaveOccurred())

		err = client.Get(context.TODO(), actualParent)
		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())

		// child should get garbage collected
		Eventually(func() error {
			err = client.Get(context.TODO(), actualParent)
			return err
		}).Should(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())

	})
})
