Feature: Secret Scoped Annotations

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Copy a single key from the related Secret into the binding secret using olm descriptor
        Given The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
          name: backends.stable.example.com
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
        And The Custom Resource is present
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            capabilities: Basic Install
          name: backend-operator.v0.1.0
        spec:
          customresourcedefinitions:
            owned:
            - description: Backend is the Schema for the backend API
              kind: Backend
              name: backends.stable.example.com
              version: v1
              specDescriptors:
                - description: Host address
                  displayName: Host address
                  path: host
                  x-descriptors:
                    - service.binding:host
              statusDescriptors:
                  - description: db credentials
                    displayName: db credentials
                    path: data.dbCredentials
                    x-descriptors:
                        - urn:alm:descriptor:io.kubernetes:Secret
                        - service.binding:username:sourceValue=username
          displayName: Backend Operator
          install:
            spec:
              deployments:
              - name: backend-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: backend-operator
                  strategy: {}
                  template:
                    metadata:
                      labels:
                        name: backend-operator
                    spec:
                      containers:
                      - command:
                        - backend-operator
                        env:
                        - name: WATCH_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['olm.targetNamespaces']
                        - name: POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        - name: OPERATOR_NAME
                          value: backend-operator
                        image: REPLACE_IMAGE
                        imagePullPolicy: Always
                        name: backend-operator
                        resources: {}
            strategy: deployment
        """
        And The Secret is present
        """
        apiVersion: v1
        kind: Secret
        metadata:
            name: ssd-1-secret
        stringData:
            username: AzureDiamond
        """
        And The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: ssd-1-service
        spec:
            host: example.com
        status:
            data:
                dbCredentials: ssd-1-secret
        """
        When Service Binding is applied
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ServiceBinding
        metadata:
            name: ssd-1
        spec:
            services:
              - group: stable.example.com
                version: v1
                kind: Backend
                name: ssd-1-service
        """
        Then Secret "ssd-1" contains "BACKEND_USERNAME" key with value "AzureDiamond"
        And Secret "ssd-1" contains "BACKEND_HOST" key with value "example.com"

    @disabled
    Scenario: Copy all keys from the Secret related to the Service resource to the binding secret with olm descriptors
        Given The Custom Resource Definition is present
    """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
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
        And The Custom Resource is present
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ClusterServiceVersion
        metadata:
          annotations:
            capabilities: Basic Install
          name: backend-operator.v0.1.0
        spec:
          customresourcedefinitions:
            owned:
            - description: Backend is the Schema for the backend API
              kind: Backend
              name: backends.stable.example.com
              version: v1
              specDescriptors:
                - description: Host address
                  displayName: Host address
                  path: host
                  x-descriptors:
                    - service.binding:host
              statusDescriptors:
                  - description: db credentials
                    displayName: db credentials
                    path: data.dbCredentials
                    x-descriptors:
                        - urn:alm:descriptor:io.kubernetes:Secret
                        - service.binding:elementType=map
          displayName: Backend Operator
          install:
            spec:
              deployments:
              - name: backend-operator
                spec:
                  replicas: 1
                  selector:
                    matchLabels:
                      name: backend-operator
                  strategy: {}
                  template:
                    metadata:
                      labels:
                        name: backend-operator
                    spec:
                      containers:
                      - command:
                        - backend-operator
                        env:
                        - name: WATCH_NAMESPACE
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.annotations['olm.targetNamespaces']
                        - name: POD_NAME
                          valueFrom:
                            fieldRef:
                              fieldPath: metadata.name
                        - name: OPERATOR_NAME
                          value: backend-operator
                        image: REPLACE_IMAGE
                        imagePullPolicy: Always
                        name: backend-operator
                        resources: {}
            strategy: deployment
        """
        And The Secret is present
        """
        apiVersion: v1
        kind: Secret
        metadata:
            name: ssd-2-secret
        stringData:
            username: AzureDiamond
            password: hunter2
        """
        And The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: ssd-2-service
        spec:
            image: docker.io/postgres
            imageName: postgres
            dbName: db-demo
        status:
            data:
                dbCredentials: ssd-2-secret
        """
        When Service Binding is applied
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ServiceBinding
        metadata:
            name: ssd-2
        spec:
            services:
              - group: stable.example.com
                version: v1
                kind: Backend
                name: ssd-2-service
        """
        Then Secret "ssd-2" contains "BACKEND_USERNAME" key with value "AzureDiamond"
        And Secret "ssd-2" contains "BACKEND_PASSWORD" key with value "hunter2"
