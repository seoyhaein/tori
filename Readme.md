## 설치 필요.
- https://github.com/mattn/go-sqlite3
- DDL 설정 필요. notion 참고.


## old

# datablock (Packrat 라고 이름 바꾸는 것을 생각해보자. blocky 는 이미 있음. 그냥 우리 고양이 이름 tori 로 하였음.)

## dependencies
~~- fsnotify (사용안함.)~~ backup 참조
~~- install : go get -u github.com/fsnotify/fsnotify~~
~~- golang.org/x/sys 이것도 자동으로 설치됨.~~
~~- fsnotify v1.8.0, x/sys v0.26.0~~

## TODO
~~- main 에서 부터 이제 어떻게 다시 시나리오를 만들어 갈지 구상 해야함.~~
- 검색 기능 넣고, grpc 연동 진행.
- 기초 grpc 넣어두고 grpc 프로젝트 만들고 고도화 함. 시작.
- config 에 grpc 관련 부수정보 넣을 것.
- sql 구문 관련해서 보안이나 여러 문제 들에 대해서 한번 체크하고 가자.
- https://github.com/golang/sync/tree/master/singleflight 이거 적용해볼 것을 생각해보자.
- golang.org/x/sync/singleflight
  ~~- grpc 컨테이너 오류 수정해줘야 함.~~  
  ~~- Dockerfile 만들었으며 테스트 진행해야함.~~  
  ~~- 사용자 편의성 생각할 것. exit 을 넣으면 종료 되는데 이게 로그가 올라오면 사라짐.(필요없음)~~
  ~~- Makefile 현재 설정이 내 노트북으로 되어 있는데 이거 확장가능하도록 하자. 우선 순위 낮음. 컨테이너로 하는 방향으로 해서 볼륨 연결해서 pb.go 파일들만 얻는 방식으로 진행하자.~~

## 수정할 것들.
~~- 일단 굴러가게만 하자.~~  
~~- invalid 해줘야 함.~~  
~~- 특정파일에 동일 규칙의 파일들이 있는지 검사해야 할까?~~  
~~- 하나의 디렉토리에 하나의 블럭만 존재하게 해야 하는가? 복수개도 존재할 수 있도록 해야 하지 않은가?~~  
~~- 디렉토리안에서 여러개의 블럭이 존재할 수 있다고 보는데. 이건 정신이 맑아지면 컨디션 좋을대 살펴보자.~~  
~~- 디렉토리는 이름이 unique 함 따라서 이것은 파일명을 잡는데 중요한 기준점임.~~  
~~- 정신이 없어서 테스트 코드 살펴봐야 한다.~~  
~~- 필터링하는 거 해줘야 함.~~
~~- sql 구문 에러 표시되는거 확인하자.~~
- 폴더에 파일이 없으면 에러 남. 이건 에러로 처리할지 생각할지 아니면 수정할지 생각해봐야 할듯. (일단 이거 빨리 처리하자)

## 생각해봐야 할 것
~~- 조금더 사용자 친화적으로 할 수 있는 방법이 없는지 생각해보자.(필요없음.)~~ 사용자 명령은 감추어 둘 것임. 어드민으로)    
~~- 디렉토리 관련해서 개발을 시작해야 할 것 같다.~~
- 사용자에게 직접 입력 받는 것은 최소화 하고, 부모 app 에게 명령을 받는 형태로 하는게 좋을 것 같다.
- 최초 사용과 그 이후 사용에 대해서 구별을 자동으로 해주는 방향으로 가자.
- github action 에 대해서 스터디 진행하자.
- 시간날때 lint 최적화 하자.

## update
- 로그 관련 표준 정하자.


### 참고 하기
https://github.com/charmbracelet/bubbletea/tree/main

### 지우기
https://v.daum.net/v/20250222170330596

###
- Golang 버전 최신 버전으로 맞출 것
- protoc version v26.1
- protoc-gen-go v1.33.0
- protoc-gen-go-grpc v1.3.0

### 확인사항
- gogoproto(https://github.com/cosmos/gogoproto) 사용 안함.
- gogoproto 사용했을때는 ~.pb.go 파일 하나만 사용하면 되었지만 표준방식으로 사용하면 두개를 만들어야 한다.
- 일단 성능적으로 낫다고 하지만, 안정적으로 standard proto 를 사용하기로 함.  
  ~~- pb 파일 생성은 window/linux 둘다 작성 함.~~ (Makefile 만들었음.)

### 설치사항
- protoc 설치
- protoc 의 설치는 https://github.com/protocolbuffers/protobuf/releases/tag/v26.1 이 링크에서 다운
- protoc 의 경우 윈도우는 설치했고, 리눅스 설치해야 함. 버전에 약간 혼동이 있었음.  일단 윈도우 버전하고 비교해보자. 이슈 발생할 듯.
- protoc-gen-go 설치
- 리눅스의 경우(ubutnu) /usr/local/bin/ 에 넣어둠.
- https://github.com/protocolbuffers/protobuf-go/releases/tag/v1.33.0 이 링크에서 설치
- 아래 grpc 설치
- go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
- 일단 go install 을 하면 gopath 기준으로 gopath/bin 에 설치가 된다. 나는 여기서 설치된것을 가져와서 해당 디렉토리에 넣어 두었다.
- 물론, 별도로 위에처럼 링크를 찾아서 해도 된다. https://github.com/grpc/grpc-go/tree/master/cmd/protoc-gen-go-grpc
- 이건 윈도우 리눅스 버전에 대한 부분을 명확히 해야 할듯하다.
- exe 확장자는 윈도우 버전이다.
- 윈도우와 리눅스를 동일하게 작성하게 만들었다.
- grpcurl 은 1.9.1 최신 버전을 설치 하였다.(https://github.com/fullstorydev/grpcurl/releases)
- grpc 다른 proto 파일들도 추가, 리눅스 버전도 해줘야 함.