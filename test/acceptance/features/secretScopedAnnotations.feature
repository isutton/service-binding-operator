Feature: Secret Scoped Annotations

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Copy a single key from the related Secret into the binding secret
        Given OLM Operator "backend" is running
        And The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
            annotations:
                service.binding/username: path={.status.data.dbCredentials},objectType=Secret,valueKey=username
        spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
                plural: backends
                singular: backend
                kind: Backend
                shortNames:
                  - bk
        """
        And The Secret is present
        """
        apiVersion: v1
        kind: Secret
        metadata:
            name: ssa-1-secret
        stringData:
            username: AzureDiamond
        """
        And The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: ssa-1-service
        spec:
            image: docker.io/postgres
            imageName: postgres
            dbName: db-demo
        status:
            data:
                dbCredentials: ssa-1-secret
        """
        When Service Binding is applied
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ServiceBinding
        metadata:
            name: ssa-1
        spec:
            services:
              - group: stable.example.com
                version: v1
                kind: Backend
                name: ssa-1-service
        """
        Then Secret "ssa-1" contains "BACKEND_USERNAME" key with value "AzureDiamond"

    Scenario: Copy all keys from the Secret related to the Service resource to the binding secret
        Given OLM Operator "backend" is running
        And The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
            annotations:
                service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
        spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
                plural: backends
                singular: backend
                kind: Backend
                shortNames:
                  - bk
        """
        And The Secret is present
        """
        apiVersion: v1
        kind: Secret
        metadata:
            name: ssa-2-secret
        stringData:
            username: AzureDiamond
            password: hunter2
        """
        And The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: ssa-2-service
        spec:
            image: docker.io/postgres
            imageName: postgres
            dbName: db-demo
        status:
            data:
                dbCredentials: ssa-2-secret
        """
        When Service Binding is applied
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ServiceBinding
        metadata:
            name: ssa-2
        spec:
            services:
              - group: stable.example.com
                version: v1
                kind: Backend
                name: ssa-2-service
        """
        Then Secret "ssa-2" contains "BACKEND_USERNAME" key with value "AzureDiamond"
        And Secret "ssa-2" contains "BACKEND_PASSWORD" key with value "hunter2"
