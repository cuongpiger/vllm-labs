- Install `vllm`:
  ```bash
  pip install vllm jinja2 --upgrade
  ```

- To list all LLM models that the vLLM is hosting, you can use the following command:
  ```bash
  export HOST="__PUT_YOUR_HOST__"
  curl http://${HOST}/v1/models
  ```
  Response:
  ```json
  {
      "object": "list",
      "data": [
          {
              "id": "Meta-Llama-3.1-8B-Instruct",
              "object": "model",
              "created": 1737783302,
              "owned_by": "vllm",
              "root": "Meta-Llama-3.1-8B-Instruct",
              "parent": null,
              "max_model_len": 131072,
              "permission": [
                  {
                      "id": "modelperm-5064860cf6514300be8194dccb4d329b",
                      "object": "model_permission",
                      "created": 1737783302,
                      "allow_create_engine": false,
                      "allow_sampling": true,
                      "allow_logprobs": true,
                      "allow_search_indices": false,
                      "allow_view": true,
                      "allow_fine_tuning": false,
                      "organization": "*",
                      "group": null,
                      "is_blocking": false
                  }
              ]
          }
      ]
  }
  ```

- Test the completion feature:
  ```bash
  curl --location 'http://${HOST}/v1/completions' \
  --header 'Content-Type: application/json' \
  --data '{
      "model": "Meta-Llama-3.1-8B-Instruct",
      "prompt": "Làm sao để nói lời chia tay",
      "max_tokens": 200,
      "temperature": 0
  }'
  ```
  Response:
  ```json
  {
      "id": "cmpl-d9e1014ef4e543828be46d45d273b83d",
      "object": "text_completion",
      "created": 1737783436,
      "model": "Meta-Llama-3.1-8B-Instruct",
      "choices": [
          {
              "index": 0,
              "text": " với người yêu cũ?\nLời chia tay có thể là một trong những điều khó khăn nhất mà bạn phải đối mặt trong cuộc sống. Nó đòi hỏi sự kiên nhẫn, sự nhạy cảm và sự tự tin. Dưới đây là một số lời khuyên để giúp bạn nói lời chia tay với người yêu cũ một cách hiệu quả:\n1. **Lựa chọn thời điểm và địa điểm phù hợp**: Hãy chọn một thời điểm và địa điểm riêng tư, thoải mái và không bị phân tâm. Điều này sẽ giúp bạn tập trung vào cuộc trò chuyện và tránh những cãi vã không cần thiết.\n2. **Lựa chọn thời điểm phù hợp**: Hãy chọn một thời điểm khi cả hai đều có thời gian và tâm trạng tốt. Tránh nói lời chia tay khi một trong hai đang bận rộn hoặc có vấn đề khác.\n3. **Lựa chọn địa điểm phù hợp**: Hãy chọn một địa điểm mà cả hai đều cảm thấy thoải mái và",
              "logprobs": null,
              "finish_reason": "length",
              "stop_reason": null,
              "prompt_logprobs": null
          }
      ],
      "usage": {
          "prompt_tokens": 9,
          "total_tokens": 209,
          "completion_tokens": 200,
          "prompt_tokens_details": null
      }
  }
  ```

- Chat with the model:
  ```bash
  curl --location 'http://${HOST}/v1/chat/completions' \
  --header 'Content-Type: application/json' \
  --data '{
      "model": "Meta-Llama-3.1-8B-Instruct",
      "messages": [
          {
              "role": "system",
              "content": "Bạn là một trợ lý ảo."
          },
          {
              "role": "user",
              "content": "Làm sao để nói lời chia tay mà không tổn thương đối phương."
          }
      ]
  }'
  ```
  Response:
  ```json
  {
      "id": "chatcmpl-259c31291b9e49648214702daaccb82e",
      "object": "chat.completion",
      "created": 1737783186,
      "model": "Meta-Llama-3.1-8B-Instruct",
      "choices": [
          {
              "index": 0,
              "message": {
                  "role": "assistant",
                  "content": "Nói lời chia tay có thể là một trong những điều khó khăn nhất anyone phải đối mặt. Dưới đây là một số gợi ý để bạn nói lời chia tay một cách nhạy cảm và không tổn thương đối phương:\n\n1. **Luôn cho phép đối phương nói cảm xúc của họ**: Cho phép đối phương nhắc đến cảm xúc của họ và lắng nghe họ. Điều này cải thiện nắm bắt và sự đối xử.\n2. **Xác định lý do riêng tư và chữa bệnh**: Nếu bạn tập trung vào lý do riêng tư và chữa bệnh, điều này làm giảm việc gây thương tổn đối với người kia.\n3. **Hãy nhẹ nhàng và cảm động**: Sử dụng lời lẽ nhẹ nhàng và cảm động để giúp đối phương cảm thấy thoải mái hơn.\n4. **Luôn có sự tôn trọng**: Trọng, cho phép và tôn trọng cảm xúc và lựa chọn của đối phương.\n5. **Kể về ký ức đẹp đẽ**: Dành thời gian chăm sóc đến những ký ức đẹp đẽ mà bạn chia sẻ cùng hướng đến người ta.\n6. **Chuẩn bị kỹ để tìm ra cách nói**\n7. **Hành động**: Hơn là có nói, giao động phê của bạn với hành động-With hành động của bạn gửi về viêm moi trùng đến người iOS.\n\nDưới đây là kinh điển trong tình yêu sau đây và cách bình tĩnh nữa,chậm lại tinh thần ổn chí:\n \n **\"Tôi yêu bạn hạnh phúc. Tôi muốn đau bị trở lại rằng tôi bỏ qua sự chê bai bạn tìm đến để sinh sống, yêu thương lại cuối cùng nhưng đáng niềm tin giữa chúng ta me miser hai người khác nhau\"time được sắp xếp và hai dựng, tôi xin lỗi và biết về từ được tổn Thương rất    Romantic kèm chú ý,ki ' ất yìm của tôi đến với đời sống đời KD,eague:(..\n\nChúc bạn thành công trong việc nói lời chia tay một cách nhạy cảm và không tổn thương đối phương.",
                  "tool_calls": []
              },
              "logprobs": null,
              "finish_reason": "stop",
              "stop_reason": null
          }
      ],
      "usage": {
          "prompt_tokens": 58,
          "total_tokens": 475,
          "completion_tokens": 417,
          "prompt_tokens_details": null
      },
      "prompt_logprobs": null
  }
  ```